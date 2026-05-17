package data

import (
	"context"
	"time"

	"gitlab.calendaria.team/services/agents/ent"
	"gitlab.calendaria.team/services/agents/ent/enum"
	"gitlab.calendaria.team/services/agents/ent/route"
	"gitlab.calendaria.team/services/agents/ent/routepoint"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type RoutesRepo interface {
	Create(ctx context.Context, dto RouteDto, points []RoutePointDto) (*ent.Route, error)
	Get(ctx context.Context, tenantID, id int64) (*ent.Route, error)
	List(ctx context.Context, tenantID int64, agentID int64, date string, status string, paginate *utils_v1.PaginateRequest) ([]*ent.Route, error)
	Count(ctx context.Context, tenantID int64, agentID int64, date string, status string) (int32, error)
	UpdatePointVisitID(ctx context.Context, pointID, visitID int64) error
	GetPoint(ctx context.Context, pointID int64) (*ent.RoutePoint, error)
	CountPointsByRoutes(ctx context.Context, routeIDs []int64) (int32, error)
	CountCompletedPointsByRoutes(ctx context.Context, routeIDs []int64) (int32, error)
	ListRouteIDsByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) ([]int64, error)
}

type routesRepo struct {
	db *ent.Client
}

func NewRoutesRepo(d *Data) RoutesRepo {
	return &routesRepo{db: d.db}
}

func (r *routesRepo) Create(ctx context.Context, dto RouteDto, points []RoutePointDto) (*ent.Route, error) {
	rt, err := r.db.Route.Create().
		SetTenantID(dto.TenantID).
		SetAgentID(dto.AgentID).
		SetDate(dto.Date).
		SetStatus(dto.Status).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	bulk := make([]*ent.RoutePointCreate, len(points))
	for i, p := range points {
		bulk[i] = r.db.RoutePoint.Create().
			SetRouteID(rt.ID).
			SetStoreTenantID(p.StoreTenantID).
			SetOrderNum(p.OrderNum)
	}
	_, err = r.db.RoutePoint.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.db.Route.Query().
		Where(route.ID(rt.ID)).
		WithPoints(func(q *ent.RoutePointQuery) {
			q.Order(ent.Asc(routepoint.FieldOrderNum))
		}).
		Only(ctx)
}

func (r *routesRepo) Get(ctx context.Context, tenantID, id int64) (*ent.Route, error) {
	return r.db.Route.Query().
		Where(route.ID(id), route.TenantID(tenantID)).
		WithPoints(func(q *ent.RoutePointQuery) {
			q.Order(ent.Asc(routepoint.FieldOrderNum))
		}).
		Only(ctx)
}

func (r *routesRepo) List(ctx context.Context, tenantID int64, agentID int64, date string, status string, paginate *utils_v1.PaginateRequest) ([]*ent.Route, error) {
	query := r.db.Route.Query().Where(route.TenantID(tenantID))

	if agentID != 0 {
		query.Where(route.AgentID(agentID))
	}
	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			query.Where(route.DateGTE(t), route.DateLT(t.AddDate(0, 0, 1)))
		}
	}
	if status != "" {
		query.Where(route.StatusEQ(enum.RouteStatus(status)))
	}
	if paginate.GetFromId() != 0 {
		query.Where(route.IDGT(paginate.GetFromId()))
	}

	limit := int(paginate.GetLimit())
	if limit == 0 {
		limit = 100
	}

	return query.
		WithPoints(func(q *ent.RoutePointQuery) {
			q.Order(ent.Asc(routepoint.FieldOrderNum))
		}).
		Limit(limit).
		Order(ent.Asc(route.FieldID)).
		All(ctx)
}

func (r *routesRepo) Count(ctx context.Context, tenantID int64, agentID int64, date string, status string) (int32, error) {
	query := r.db.Route.Query().Where(route.TenantID(tenantID))

	if agentID != 0 {
		query.Where(route.AgentID(agentID))
	}
	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err == nil {
			query.Where(route.DateGTE(t), route.DateLT(t.AddDate(0, 0, 1)))
		}
	}
	if status != "" {
		query.Where(route.StatusEQ(enum.RouteStatus(status)))
	}

	count, err := query.Count(ctx)
	return int32(count), err
}

func (r *routesRepo) UpdatePointVisitID(ctx context.Context, pointID, visitID int64) error {
	return r.db.RoutePoint.UpdateOneID(pointID).SetVisitID(visitID).Exec(ctx)
}

func (r *routesRepo) GetPoint(ctx context.Context, pointID int64) (*ent.RoutePoint, error) {
	return r.db.RoutePoint.Query().
		Where(routepoint.ID(pointID)).
		WithRoute().
		Only(ctx)
}

func (r *routesRepo) CountPointsByRoutes(ctx context.Context, routeIDs []int64) (int32, error) {
	count, err := r.db.RoutePoint.Query().
		Where(routepoint.RouteIDIn(routeIDs...)).
		Count(ctx)
	return int32(count), err
}

func (r *routesRepo) CountCompletedPointsByRoutes(ctx context.Context, routeIDs []int64) (int32, error) {
	count, err := r.db.RoutePoint.Query().
		Where(routepoint.RouteIDIn(routeIDs...), routepoint.VisitIDGT(0)).
		Count(ctx)
	return int32(count), err
}

func (r *routesRepo) ListRouteIDsByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) ([]int64, error) {
	return r.db.Route.Query().
		Where(
			route.TenantID(tenantID),
			route.AgentID(agentID),
			route.DateGTE(from),
			route.DateLTE(to),
		).
		IDs(ctx)
}
