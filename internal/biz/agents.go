package biz

import (
	"context"
	"fmt"
	"time"

	"gitlab.calendaria.team/services/agents/ent"
	"gitlab.calendaria.team/services/agents/ent/enum"
	"gitlab.calendaria.team/services/agents/internal/data"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

// TenantsClient is a stub interface for cross-service tenant creation (Story 6.4).
type TenantsClient interface {
	CreateTenant(ctx context.Context, storeName, ownerPhone string) (int64, error)
}

type stubTenantsClient struct{}

func (s *stubTenantsClient) CreateTenant(_ context.Context, _, _ string) (int64, error) {
	// Stub: returns a fake tenant ID. Replace with real gRPC client.
	return time.Now().UnixMilli(), nil
}

type AgentsUsecase struct {
	log             *log.Helper
	routesRepo      data.RoutesRepo
	visitsRepo      data.VisitsRepo
	photosRepo      data.VisitPhotosRepo
	onboardingsRepo data.OnboardingsRepo
	tenantsClient   TenantsClient
	storesClient    data.StoresClient
}

func NewAgentsUsecase(
	logger log.Logger,
	routesRepo data.RoutesRepo,
	visitsRepo data.VisitsRepo,
	photosRepo data.VisitPhotosRepo,
	onboardingsRepo data.OnboardingsRepo,
	storesClient data.StoresClient,
) *AgentsUsecase {
	return &AgentsUsecase{
		log:             log.NewHelper(logger),
		routesRepo:      routesRepo,
		visitsRepo:      visitsRepo,
		photosRepo:      photosRepo,
		onboardingsRepo: onboardingsRepo,
		tenantsClient:   &stubTenantsClient{},
		storesClient:    storesClient,
	}
}

// --- Story 6.1: Routes ---

func (uc *AgentsUsecase) CreateRoute(ctx context.Context, dto data.RouteDto, points []data.RoutePointDto) (*ent.Route, error) {
	if dto.AgentID == 0 {
		return nil, fmt.Errorf("agent_id is required")
	}
	if dto.Date.IsZero() {
		return nil, fmt.Errorf("date is required")
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("at least one route point is required")
	}
	dto.Status = enum.Planned
	return uc.routesRepo.Create(ctx, dto, points)
}

func (uc *AgentsUsecase) GetRoute(ctx context.Context, tenantID, id int64) (*ent.Route, error) {
	return uc.routesRepo.Get(ctx, tenantID, id)
}

func (uc *AgentsUsecase) ListRoutes(ctx context.Context, tenantID, agentID int64, date, status string, paginate *utils_v1.PaginateRequest) ([]*ent.Route, int32, error) {
	items, err := uc.routesRepo.List(ctx, tenantID, agentID, date, status, paginate)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.routesRepo.Count(ctx, tenantID, agentID, date, status)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// --- Story 6.2: GPS Visits ---

func (uc *AgentsUsecase) CheckIn(ctx context.Context, tenantID int64, routePointID int64, agentID int64, lat, lon decimal.Decimal) (*ent.Visit, error) {
	// Verify route point exists
	point, err := uc.routesRepo.GetPoint(ctx, routePointID)
	if err != nil {
		return nil, fmt.Errorf("route point not found")
	}

	// Verify the route belongs to this tenant
	if point.Edges.Route != nil && point.Edges.Route.TenantID != tenantID {
		return nil, fmt.Errorf("route point not found")
	}

	// Validate proximity to store (Story 10.4)
	if point.StoreTenantID != 0 && uc.storesClient != nil {
		agentLat, _ := lat.Float64()
		agentLon, _ := lon.Float64()
		ok, err := uc.storesClient.ValidateProximity(ctx, point.StoreTenantID, agentLat, agentLon)
		if err != nil {
			uc.log.Warnf("failed to validate proximity for store %d: %v", point.StoreTenantID, err)
			// Don't block check-in on service failure
		} else if !ok {
			return nil, fmt.Errorf("agent is too far from the store (must be within 200m)")
		}
	}

	// Check if already checked in
	existing, err := uc.visitsRepo.GetByRoutePointID(ctx, routePointID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("already checked in at this route point")
	}

	now := time.Now()
	v, err := uc.visitsRepo.Create(ctx, data.VisitDto{
		TenantID:     tenantID,
		RoutePointID: routePointID,
		AgentID:      agentID,
		CheckinLat:   lat,
		CheckinLon:   lon,
		CheckinAt:    now,
	})
	if err != nil {
		return nil, err
	}

	// Link visit to route point
	_ = uc.routesRepo.UpdatePointVisitID(ctx, routePointID, v.ID)

	return v, nil
}

func (uc *AgentsUsecase) CheckOut(ctx context.Context, tenantID int64, routePointID int64, lat, lon decimal.Decimal) (*ent.Visit, error) {
	v, err := uc.visitsRepo.GetByRoutePointID(ctx, routePointID)
	if err != nil {
		return nil, fmt.Errorf("no checkin found for this route point")
	}
	if v.TenantID != tenantID {
		return nil, fmt.Errorf("no checkin found for this route point")
	}
	if v.CheckoutAt != nil {
		return nil, fmt.Errorf("already checked out")
	}

	now := time.Now()
	duration := int64(now.Sub(v.CheckinAt).Seconds())

	return uc.visitsRepo.CheckOut(ctx, v.ID, lat, lon, now, duration)
}

func (uc *AgentsUsecase) ListVisits(ctx context.Context, tenantID, agentID int64, dateFrom, dateTo string, paginate *utils_v1.PaginateRequest) ([]*ent.Visit, int32, error) {
	items, err := uc.visitsRepo.List(ctx, tenantID, agentID, dateFrom, dateTo, paginate)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.visitsRepo.Count(ctx, tenantID, agentID, dateFrom, dateTo)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// --- Story 6.3: Photo Reports ---

func (uc *AgentsUsecase) AddVisitPhoto(ctx context.Context, tenantID int64, visitID int64, mediaURL string) (*ent.VisitPhoto, error) {
	if mediaURL == "" {
		return nil, fmt.Errorf("media_url is required")
	}
	return uc.photosRepo.Create(ctx, data.VisitPhotoDto{
		VisitID:  visitID,
		MediaURL: mediaURL,
	})
}

func (uc *AgentsUsecase) GetVisitPhotos(ctx context.Context, visitID int64) ([]*ent.VisitPhoto, error) {
	return uc.photosRepo.ListByVisit(ctx, visitID)
}

// --- Story 6.4: Onboard Store ---

func (uc *AgentsUsecase) OnboardStore(ctx context.Context, tenantID, agentID int64, storeName, ownerPhone string) (int64, error) {
	if storeName == "" {
		return 0, fmt.Errorf("store_name is required")
	}
	if ownerPhone == "" {
		return 0, fmt.Errorf("owner_phone is required")
	}

	// Create tenant via external service (stub)
	newTenantID, err := uc.tenantsClient.CreateTenant(ctx, storeName, ownerPhone)
	if err != nil {
		return 0, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Record the onboarding
	_, err = uc.onboardingsRepo.Create(ctx, data.OnboardingDto{
		TenantID:    tenantID,
		AgentID:     agentID,
		NewTenantID: newTenantID,
		StoreName:   storeName,
		OwnerPhone:  ownerPhone,
	})
	if err != nil {
		uc.log.Errorf("failed to record onboarding: %v", err)
	}

	return newTenantID, nil
}

// --- Story 6.5: Agent Reports ---

func (uc *AgentsUsecase) GetAgentReport(ctx context.Context, tenantID, agentID int64, dateFrom, dateTo string) (*AgentReportResult, error) {
	from, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid date_from format")
	}
	to, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil, fmt.Errorf("invalid date_to format")
	}
	// Include the full end day
	toEnd := to.AddDate(0, 0, 1)

	visitsCount, err := uc.visitsRepo.CountByAgentAndDateRange(ctx, tenantID, agentID, from, toEnd)
	if err != nil {
		return nil, err
	}

	onboardingsCount, err := uc.onboardingsRepo.CountByAgentAndDateRange(ctx, tenantID, agentID, from, toEnd)
	if err != nil {
		return nil, err
	}

	totalDuration, err := uc.visitsRepo.TotalDurationByAgentAndDateRange(ctx, tenantID, agentID, from, toEnd)
	if err != nil {
		return nil, err
	}

	routeIDs, err := uc.routesRepo.ListRouteIDsByAgentAndDateRange(ctx, tenantID, agentID, from, to)
	if err != nil {
		return nil, err
	}

	var totalPoints, completedPoints int32
	if len(routeIDs) > 0 {
		totalPoints, err = uc.routesRepo.CountPointsByRoutes(ctx, routeIDs)
		if err != nil {
			return nil, err
		}
		completedPoints, err = uc.routesRepo.CountCompletedPointsByRoutes(ctx, routeIDs)
		if err != nil {
			return nil, err
		}
	}

	var completionPct float64
	if totalPoints > 0 {
		completionPct = float64(completedPoints) / float64(totalPoints) * 100
	}

	return &AgentReportResult{
		AgentID:              agentID,
		DateFrom:             dateFrom,
		DateTo:               dateTo,
		VisitsCount:          visitsCount,
		OrdersCount:          0, // Stub: would come from orders service
		OnboardingsCount:     onboardingsCount,
		TotalRoutePoints:     totalPoints,
		CompletedRoutePoints: completedPoints,
		RouteCompletionPct:   completionPct,
		TotalDurationSeconds: totalDuration,
	}, nil
}

type AgentReportResult struct {
	AgentID              int64
	DateFrom             string
	DateTo               string
	VisitsCount          int32
	OrdersCount          int32
	OnboardingsCount     int32
	TotalRoutePoints     int32
	CompletedRoutePoints int32
	RouteCompletionPct   float64
	TotalDurationSeconds int64
}

