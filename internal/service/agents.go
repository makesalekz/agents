package service

import (
	"context"
	"time"

	v1 "github.com/makesalekz/agents/api/agents/v1"
	"github.com/makesalekz/agents/ent"
	"github.com/makesalekz/agents/internal/biz"
	"github.com/makesalekz/agents/internal/data"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"
	"github.com/makesalekz/utils/v2/auth"

	"github.com/shopspring/decimal"
)

type AgentsService struct {
	v1.UnimplementedAgentsServiceServer

	uc *biz.AgentsUsecase
}

func NewAgentsService(uc *biz.AgentsUsecase) *AgentsService {
	return &AgentsService{uc: uc}
}

// --- Story 6.1: Routes ---

func (s *AgentsService) CreateRoute(ctx context.Context, req *v1.CreateRouteRequest) (*v1.CreateRouteReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	date, err := time.Parse("2006-01-02", req.GetDate())
	if err != nil {
		return nil, v1.ErrorInvalidRequest("invalid date format, expected YYYY-MM-DD")
	}

	points := make([]data.RoutePointDto, 0, len(req.GetPoints()))
	for _, p := range req.GetPoints() {
		points = append(points, data.RoutePointDto{
			StoreTenantID: p.GetStoreTenantId(),
			OrderNum:      p.GetOrderNum(),
		})
	}

	route, err := s.uc.CreateRoute(ctx, data.RouteDto{
		TenantID: tenantID,
		AgentID:  req.GetAgentId(),
		Date:     date,
	}, points)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("%s", err.Error())
	}

	return &v1.CreateRouteReply{Route: replyRoute(route)}, nil
}

func (s *AgentsService) GetRoute(ctx context.Context, req *v1.GetRouteRequest) (*v1.GetRouteReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	route, err := s.uc.GetRoute(ctx, tenantID, req.GetId())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorNotFound("route not found")
		}
		return nil, err
	}

	return &v1.GetRouteReply{Route: replyRoute(route)}, nil
}

func (s *AgentsService) ListRoutes(ctx context.Context, req *v1.ListRoutesRequest) (*v1.ListRoutesReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	paginate := req.GetPaginate()
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	items, total, err := s.uc.ListRoutes(ctx, tenantID, req.GetAgentId(), req.GetDate(), req.GetStatus(), paginate)
	if err != nil {
		return nil, err
	}

	routes := make([]*v1.Route, 0, len(items))
	for _, item := range items {
		routes = append(routes, replyRoute(item))
	}

	var fromID, toID *int64
	if len(items) > 0 {
		f := items[0].ID
		t := items[len(items)-1].ID
		fromID = &f
		toID = &t
	}

	return &v1.ListRoutesReply{
		Items: routes,
		Paginate: &utils_v1.PaginateReply{
			Total:  &total,
			FromId: fromID,
			ToId:   toID,
		},
	}, nil
}

// --- Story 6.2: GPS Visits ---

func (s *AgentsService) CheckIn(ctx context.Context, req *v1.CheckInRequest) (*v1.CheckInReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}
	agentID := auth.GetActorIdFromContext(ctx)

	lat := parseDecimal(req.GetLat())
	lon := parseDecimal(req.GetLon())

	visit, err := s.uc.CheckIn(ctx, tenantID, req.GetRoutePointId(), agentID, lat, lon)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("%s", err.Error())
	}

	return &v1.CheckInReply{Visit: replyVisit(visit)}, nil
}

func (s *AgentsService) CheckOut(ctx context.Context, req *v1.CheckOutRequest) (*v1.CheckOutReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	lat := parseDecimal(req.GetLat())
	lon := parseDecimal(req.GetLon())

	visit, err := s.uc.CheckOut(ctx, tenantID, req.GetRoutePointId(), lat, lon)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("%s", err.Error())
	}

	return &v1.CheckOutReply{Visit: replyVisit(visit)}, nil
}

func (s *AgentsService) ListVisits(ctx context.Context, req *v1.ListVisitsRequest) (*v1.ListVisitsReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	paginate := req.GetPaginate()
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	items, total, err := s.uc.ListVisits(ctx, tenantID, req.GetAgentId(), req.GetDateFrom(), req.GetDateTo(), paginate)
	if err != nil {
		return nil, err
	}

	visits := make([]*v1.Visit, 0, len(items))
	for _, item := range items {
		visits = append(visits, replyVisit(item))
	}

	var fromID, toID *int64
	if len(items) > 0 {
		f := items[0].ID
		t := items[len(items)-1].ID
		fromID = &f
		toID = &t
	}

	return &v1.ListVisitsReply{
		Items: visits,
		Paginate: &utils_v1.PaginateReply{
			Total:  &total,
			FromId: fromID,
			ToId:   toID,
		},
	}, nil
}

// --- Story 6.3: Photo Reports ---

func (s *AgentsService) AddVisitPhoto(ctx context.Context, req *v1.AddVisitPhotoRequest) (*v1.AddVisitPhotoReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	photo, err := s.uc.AddVisitPhoto(ctx, tenantID, req.GetVisitId(), req.GetMediaUrl())
	if err != nil {
		return nil, v1.ErrorInvalidRequest("%s", err.Error())
	}

	return &v1.AddVisitPhotoReply{Photo: replyVisitPhoto(photo)}, nil
}

func (s *AgentsService) GetVisitPhotos(ctx context.Context, req *v1.GetVisitPhotosRequest) (*v1.GetVisitPhotosReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	items, err := s.uc.GetVisitPhotos(ctx, req.GetVisitId())
	if err != nil {
		return nil, err
	}

	photos := make([]*v1.VisitPhoto, 0, len(items))
	for _, item := range items {
		photos = append(photos, replyVisitPhoto(item))
	}

	return &v1.GetVisitPhotosReply{Items: photos}, nil
}

// --- Story 6.4: Onboard Store ---

func (s *AgentsService) OnboardStore(ctx context.Context, req *v1.OnboardStoreRequest) (*v1.OnboardStoreReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}
	agentID := auth.GetActorIdFromContext(ctx)

	newTenantID, err := s.uc.OnboardStore(ctx, tenantID, agentID, req.GetStoreName(), req.GetOwnerPhone())
	if err != nil {
		return nil, v1.ErrorInvalidRequest("%s", err.Error())
	}

	return &v1.OnboardStoreReply{NewTenantId: newTenantID}, nil
}

// --- Story 6.5: Agent Reports ---

func (s *AgentsService) GetAgentReport(ctx context.Context, req *v1.GetAgentReportRequest) (*v1.GetAgentReportReply, error) {
	tenantID := auth.GetTenantIdFromContext(ctx)
	if tenantID == 0 {
		return nil, v1.ErrorInvalidRequest("empty tenant id")
	}

	report, err := s.uc.GetAgentReport(ctx, tenantID, req.GetAgentId(), req.GetDateFrom(), req.GetDateTo())
	if err != nil {
		return nil, v1.ErrorInvalidRequest("%s", err.Error())
	}

	return &v1.GetAgentReportReply{
		Report: &v1.AgentReport{
			AgentId:              report.AgentID,
			DateFrom:             report.DateFrom,
			DateTo:               report.DateTo,
			VisitsCount:          report.VisitsCount,
			OrdersCount:          report.OrdersCount,
			OnboardingsCount:     report.OnboardingsCount,
			TotalRoutePoints:     report.TotalRoutePoints,
			CompletedRoutePoints: report.CompletedRoutePoints,
			RouteCompletionPct:   report.RouteCompletionPct,
			TotalDurationSeconds: report.TotalDurationSeconds,
		},
	}, nil
}

// --- Reply helpers ---

func replyRoute(r *ent.Route) *v1.Route {
	resp := &v1.Route{
		Id:        r.ID,
		TenantId:  r.TenantID,
		AgentId:   r.AgentID,
		Date:      r.Date.Format("2006-01-02"),
		Status:    string(r.Status),
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	for _, p := range r.Edges.Points {
		resp.Points = append(resp.Points, replyRoutePoint(p))
	}
	return resp
}

func replyRoutePoint(p *ent.RoutePoint) *v1.RoutePoint {
	return &v1.RoutePoint{
		Id:            p.ID,
		RouteId:       p.RouteID,
		StoreTenantId: p.StoreTenantID,
		OrderNum:      p.OrderNum,
		VisitId:       p.VisitID,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func replyVisit(v *ent.Visit) *v1.Visit {
	resp := &v1.Visit{
		Id:              v.ID,
		TenantId:        v.TenantID,
		RoutePointId:    v.RoutePointID,
		AgentId:         v.AgentID,
		CheckinLat:      v.CheckinLat.String(),
		CheckinLon:      v.CheckinLon.String(),
		CheckinAt:       v.CheckinAt.Format("2006-01-02T15:04:05Z07:00"),
		DurationSeconds: v.DurationSeconds,
		CreatedAt:       v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if v.CheckoutAt != nil {
		resp.CheckoutLat = v.CheckoutLat.String()
		resp.CheckoutLon = v.CheckoutLon.String()
		resp.CheckoutAt = v.CheckoutAt.Format("2006-01-02T15:04:05Z07:00")
	}
	for _, p := range v.Edges.Photos {
		resp.Photos = append(resp.Photos, replyVisitPhoto(p))
	}
	return resp
}

func replyVisitPhoto(p *ent.VisitPhoto) *v1.VisitPhoto {
	return &v1.VisitPhoto{
		Id:        p.ID,
		VisitId:   p.VisitID,
		MediaUrl:  p.MediaURL,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func parseDecimal(s string) decimal.Decimal {
	d, _ := decimal.NewFromString(s)
	return d
}
