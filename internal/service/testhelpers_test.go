package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"gitlab.calendaria.team/services/agents/ent"
	"gitlab.calendaria.team/services/agents/ent/enum"
	"gitlab.calendaria.team/services/agents/internal/biz"
	"gitlab.calendaria.team/services/agents/internal/data"
	stores_v1 "gitlab.calendaria.team/services/stores/api/stores/v1"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

var errNotFound = errors.New("not found")

// --- Mock RoutesRepo ---

type mockRoutesRepo struct {
	routes     map[int64]*ent.Route
	points     map[int64]*ent.RoutePoint
	nextRouteID int64
	nextPointID int64
}

func newMockRoutesRepo() *mockRoutesRepo {
	return &mockRoutesRepo{
		routes:      make(map[int64]*ent.Route),
		points:      make(map[int64]*ent.RoutePoint),
		nextRouteID: 1,
		nextPointID: 1,
	}
}

func (m *mockRoutesRepo) Create(_ context.Context, dto data.RouteDto, pointDtos []data.RoutePointDto) (*ent.Route, error) {
	r := &ent.Route{
		ID:        m.nextRouteID,
		TenantID:  dto.TenantID,
		AgentID:   dto.AgentID,
		Date:      dto.Date,
		Status:    dto.Status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.routes[m.nextRouteID] = r
	m.nextRouteID++

	var pts []*ent.RoutePoint
	for _, pd := range pointDtos {
		pt := &ent.RoutePoint{
			ID:            m.nextPointID,
			RouteID:       r.ID,
			StoreTenantID: pd.StoreTenantID,
			OrderNum:      pd.OrderNum,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		m.points[m.nextPointID] = pt
		pts = append(pts, pt)
		m.nextPointID++
	}
	r.Edges.Points = pts

	return r, nil
}

func (m *mockRoutesRepo) Get(_ context.Context, tenantID, id int64) (*ent.Route, error) {
	r, ok := m.routes[id]
	if !ok || r.TenantID != tenantID {
		return nil, errNotFound
	}
	return r, nil
}

func (m *mockRoutesRepo) List(_ context.Context, tenantID int64, agentID int64, date string, status string, paginate *utils_v1.PaginateRequest) ([]*ent.Route, error) {
	var result []*ent.Route
	for _, r := range m.routes {
		if r.TenantID != tenantID {
			continue
		}
		if agentID != 0 && r.AgentID != agentID {
			continue
		}
		if date != "" {
			if t, err := time.Parse("2006-01-02", date); err == nil {
				if !r.Date.Equal(t) {
					continue
				}
			}
		}
		if status != "" && string(r.Status) != status {
			continue
		}
		if paginate.GetFromId() != 0 && r.ID <= paginate.GetFromId() {
			continue
		}
		result = append(result, r)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	limit := int(paginate.GetLimit())
	if limit == 0 {
		limit = 100
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRoutesRepo) Count(_ context.Context, tenantID int64, agentID int64, date string, status string) (int32, error) {
	var count int32
	for _, r := range m.routes {
		if r.TenantID != tenantID {
			continue
		}
		if agentID != 0 && r.AgentID != agentID {
			continue
		}
		if date != "" {
			if t, err := time.Parse("2006-01-02", date); err == nil {
				if !r.Date.Equal(t) {
					continue
				}
			}
		}
		if status != "" && string(r.Status) != status {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockRoutesRepo) UpdatePointVisitID(_ context.Context, pointID, visitID int64) error {
	pt, ok := m.points[pointID]
	if !ok {
		return errNotFound
	}
	pt.VisitID = visitID
	return nil
}

func (m *mockRoutesRepo) GetPoint(_ context.Context, pointID int64) (*ent.RoutePoint, error) {
	pt, ok := m.points[pointID]
	if !ok {
		return nil, errNotFound
	}
	// Attach route edge
	if r, ok := m.routes[pt.RouteID]; ok {
		pt.Edges.Route = r
	}
	return pt, nil
}

func (m *mockRoutesRepo) CountPointsByRoutes(_ context.Context, routeIDs []int64) (int32, error) {
	idSet := make(map[int64]bool)
	for _, id := range routeIDs {
		idSet[id] = true
	}
	var count int32
	for _, pt := range m.points {
		if idSet[pt.RouteID] {
			count++
		}
	}
	return count, nil
}

func (m *mockRoutesRepo) CountCompletedPointsByRoutes(_ context.Context, routeIDs []int64) (int32, error) {
	idSet := make(map[int64]bool)
	for _, id := range routeIDs {
		idSet[id] = true
	}
	var count int32
	for _, pt := range m.points {
		if idSet[pt.RouteID] && pt.VisitID > 0 {
			count++
		}
	}
	return count, nil
}

func (m *mockRoutesRepo) ListRouteIDsByAgentAndDateRange(_ context.Context, tenantID, agentID int64, from, to time.Time) ([]int64, error) {
	var ids []int64
	for _, r := range m.routes {
		if r.TenantID == tenantID && r.AgentID == agentID &&
			!r.Date.Before(from) && !r.Date.After(to) {
			ids = append(ids, r.ID)
		}
	}
	return ids, nil
}

// --- Mock VisitsRepo ---

type mockVisitsRepo struct {
	visits  map[int64]*ent.Visit
	nextID  int64
}

func newMockVisitsRepo() *mockVisitsRepo {
	return &mockVisitsRepo{
		visits: make(map[int64]*ent.Visit),
		nextID: 1,
	}
}

func (m *mockVisitsRepo) Create(_ context.Context, dto data.VisitDto) (*ent.Visit, error) {
	v := &ent.Visit{
		ID:           m.nextID,
		TenantID:     dto.TenantID,
		RoutePointID: dto.RoutePointID,
		AgentID:      dto.AgentID,
		CheckinLat:   dto.CheckinLat,
		CheckinLon:   dto.CheckinLon,
		CheckinAt:    dto.CheckinAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.visits[m.nextID] = v
	m.nextID++
	return v, nil
}

func (m *mockVisitsRepo) GetByRoutePointID(_ context.Context, routePointID int64) (*ent.Visit, error) {
	for _, v := range m.visits {
		if v.RoutePointID == routePointID {
			return v, nil
		}
	}
	return nil, errNotFound
}

func (m *mockVisitsRepo) CheckOut(_ context.Context, id int64, lat, lon decimal.Decimal, checkoutAt time.Time, durationSeconds int64) (*ent.Visit, error) {
	v, ok := m.visits[id]
	if !ok {
		return nil, errNotFound
	}
	v.CheckoutLat = lat
	v.CheckoutLon = lon
	v.CheckoutAt = &checkoutAt
	v.DurationSeconds = durationSeconds
	v.UpdatedAt = time.Now()
	return v, nil
}

func (m *mockVisitsRepo) List(_ context.Context, tenantID, agentID int64, dateFrom, dateTo string, paginate *utils_v1.PaginateRequest) ([]*ent.Visit, error) {
	var result []*ent.Visit
	for _, v := range m.visits {
		if v.TenantID != tenantID {
			continue
		}
		if agentID != 0 && v.AgentID != agentID {
			continue
		}
		if paginate.GetFromId() != 0 && v.ID <= paginate.GetFromId() {
			continue
		}
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	limit := int(paginate.GetLimit())
	if limit == 0 {
		limit = 100
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockVisitsRepo) Count(_ context.Context, tenantID, agentID int64, dateFrom, dateTo string) (int32, error) {
	var count int32
	for _, v := range m.visits {
		if v.TenantID != tenantID {
			continue
		}
		if agentID != 0 && v.AgentID != agentID {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockVisitsRepo) CountByAgentAndDateRange(_ context.Context, tenantID, agentID int64, from, to time.Time) (int32, error) {
	var count int32
	for _, v := range m.visits {
		if v.TenantID == tenantID && v.AgentID == agentID &&
			!v.CheckinAt.Before(from) && !v.CheckinAt.After(to) {
			count++
		}
	}
	return count, nil
}

func (m *mockVisitsRepo) TotalDurationByAgentAndDateRange(_ context.Context, tenantID, agentID int64, from, to time.Time) (int64, error) {
	var total int64
	for _, v := range m.visits {
		if v.TenantID == tenantID && v.AgentID == agentID &&
			!v.CheckinAt.Before(from) && !v.CheckinAt.After(to) {
			total += v.DurationSeconds
		}
	}
	return total, nil
}

// --- Mock VisitPhotosRepo ---

type mockVisitPhotosRepo struct {
	photos map[int64]*ent.VisitPhoto
	nextID int64
}

func newMockVisitPhotosRepo() *mockVisitPhotosRepo {
	return &mockVisitPhotosRepo{
		photos: make(map[int64]*ent.VisitPhoto),
		nextID: 1,
	}
}

func (m *mockVisitPhotosRepo) Create(_ context.Context, dto data.VisitPhotoDto) (*ent.VisitPhoto, error) {
	p := &ent.VisitPhoto{
		ID:        m.nextID,
		VisitID:   dto.VisitID,
		MediaURL:  dto.MediaURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.photos[m.nextID] = p
	m.nextID++
	return p, nil
}

func (m *mockVisitPhotosRepo) ListByVisit(_ context.Context, visitID int64) ([]*ent.VisitPhoto, error) {
	var result []*ent.VisitPhoto
	for _, p := range m.photos {
		if p.VisitID == visitID {
			result = append(result, p)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

// --- Mock OnboardingsRepo ---

type mockOnboardingsRepo struct {
	onboardings map[int64]*ent.Onboarding
	nextID      int64
}

func newMockOnboardingsRepo() *mockOnboardingsRepo {
	return &mockOnboardingsRepo{
		onboardings: make(map[int64]*ent.Onboarding),
		nextID:      1,
	}
}

func (m *mockOnboardingsRepo) Create(_ context.Context, dto data.OnboardingDto) (*ent.Onboarding, error) {
	o := &ent.Onboarding{
		ID:          m.nextID,
		TenantID:    dto.TenantID,
		AgentID:     dto.AgentID,
		NewTenantID: dto.NewTenantID,
		StoreName:   dto.StoreName,
		OwnerPhone:  dto.OwnerPhone,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.onboardings[m.nextID] = o
	m.nextID++
	return o, nil
}

func (m *mockOnboardingsRepo) CountByAgentAndDateRange(_ context.Context, tenantID, agentID int64, from, to time.Time) (int32, error) {
	var count int32
	for _, o := range m.onboardings {
		if o.TenantID == tenantID && o.AgentID == agentID &&
			!o.CreatedAt.Before(from) && !o.CreatedAt.After(to) {
			count++
		}
	}
	return count, nil
}

// --- Mock StoresClient ---

type mockStoresClient struct{}

func (m *mockStoresClient) GetStore(_ context.Context, _ int64) (*stores_v1.Store, error) {
	return &stores_v1.Store{}, nil
}

func (m *mockStoresClient) ValidateProximity(_ context.Context, _ int64, _, _ float64) (bool, error) {
	return true, nil // Always allow in tests
}

// --- Test setup ---

func setupService() (*AgentsService, *mockRoutesRepo, *mockVisitsRepo, *mockVisitPhotosRepo, *mockOnboardingsRepo) {
	routesRepo := newMockRoutesRepo()
	visitsRepo := newMockVisitsRepo()
	photosRepo := newMockVisitPhotosRepo()
	onboardingsRepo := newMockOnboardingsRepo()
	uc := biz.NewAgentsUsecase(log.DefaultLogger, routesRepo, visitsRepo, photosRepo, onboardingsRepo, &mockStoresClient{})
	svc := NewAgentsService(uc)
	return svc, routesRepo, visitsRepo, photosRepo, onboardingsRepo
}

func ctxWithTenant(tenantID int64) context.Context {
	return auth.NewTenantContext(context.Background(), tenantID)
}

func ctxWithTenantAndActor(tenantID, actorID int64) context.Context {
	ctx := auth.NewTenantContext(context.Background(), tenantID)
	return auth.NewActorContext(ctx, actorID)
}

// suppress unused warnings
var _ = enum.Planned
