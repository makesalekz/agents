package data

import (
	"context"
	"time"

	"github.com/makesalekz/agents/ent"
	"github.com/makesalekz/agents/ent/visit"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"

	"github.com/shopspring/decimal"
)

type VisitsRepo interface {
	Create(ctx context.Context, dto VisitDto) (*ent.Visit, error)
	GetByRoutePointID(ctx context.Context, routePointID int64) (*ent.Visit, error)
	CheckOut(ctx context.Context, id int64, lat, lon decimal.Decimal, checkoutAt time.Time, durationSeconds int64) (*ent.Visit, error)
	List(ctx context.Context, tenantID, agentID int64, dateFrom, dateTo string, paginate *utils_v1.PaginateRequest) ([]*ent.Visit, error)
	Count(ctx context.Context, tenantID, agentID int64, dateFrom, dateTo string) (int32, error)
	CountByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) (int32, error)
	TotalDurationByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) (int64, error)
}

type visitsRepo struct {
	db *ent.Client
}

func NewVisitsRepo(d *Data) VisitsRepo {
	return &visitsRepo{db: d.db}
}

func (r *visitsRepo) Create(ctx context.Context, dto VisitDto) (*ent.Visit, error) {
	return r.db.Visit.Create().
		SetTenantID(dto.TenantID).
		SetRoutePointID(dto.RoutePointID).
		SetAgentID(dto.AgentID).
		SetCheckinLat(dto.CheckinLat).
		SetCheckinLon(dto.CheckinLon).
		SetCheckinAt(dto.CheckinAt).
		Save(ctx)
}

func (r *visitsRepo) GetByRoutePointID(ctx context.Context, routePointID int64) (*ent.Visit, error) {
	return r.db.Visit.Query().
		Where(visit.RoutePointID(routePointID)).
		WithPhotos().
		Only(ctx)
}

func (r *visitsRepo) CheckOut(ctx context.Context, id int64, lat, lon decimal.Decimal, checkoutAt time.Time, durationSeconds int64) (*ent.Visit, error) {
	err := r.db.Visit.UpdateOneID(id).
		SetCheckoutLat(lat).
		SetCheckoutLon(lon).
		SetCheckoutAt(checkoutAt).
		SetDurationSeconds(durationSeconds).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return r.db.Visit.Query().
		Where(visit.ID(id)).
		WithPhotos().
		Only(ctx)
}

func (r *visitsRepo) listQuery(tenantID, agentID int64, dateFrom, dateTo string) *ent.VisitQuery {
	query := r.db.Visit.Query().Where(visit.TenantID(tenantID))
	if agentID != 0 {
		query.Where(visit.AgentID(agentID))
	}
	if dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			query.Where(visit.CheckinAtGTE(t))
		}
	}
	if dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			query.Where(visit.CheckinAtLT(t.AddDate(0, 0, 1)))
		}
	}
	return query
}

func (r *visitsRepo) List(ctx context.Context, tenantID, agentID int64, dateFrom, dateTo string, paginate *utils_v1.PaginateRequest) ([]*ent.Visit, error) {
	query := r.listQuery(tenantID, agentID, dateFrom, dateTo)

	if paginate.GetFromId() != 0 {
		query.Where(visit.IDGT(paginate.GetFromId()))
	}

	limit := int(paginate.GetLimit())
	if limit == 0 {
		limit = 100
	}

	return query.WithPhotos().Limit(limit).Order(ent.Asc(visit.FieldID)).All(ctx)
}

func (r *visitsRepo) Count(ctx context.Context, tenantID, agentID int64, dateFrom, dateTo string) (int32, error) {
	count, err := r.listQuery(tenantID, agentID, dateFrom, dateTo).Count(ctx)
	return int32(count), err
}

func (r *visitsRepo) CountByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) (int32, error) {
	count, err := r.db.Visit.Query().
		Where(
			visit.TenantID(tenantID),
			visit.AgentID(agentID),
			visit.CheckinAtGTE(from),
			visit.CheckinAtLTE(to),
		).
		Count(ctx)
	return int32(count), err
}

func (r *visitsRepo) TotalDurationByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) (int64, error) {
	var results []struct {
		Sum int64 `json:"sum"`
	}
	err := r.db.Visit.Query().
		Where(
			visit.TenantID(tenantID),
			visit.AgentID(agentID),
			visit.CheckinAtGTE(from),
			visit.CheckinAtLTE(to),
		).
		Aggregate(ent.Sum(visit.FieldDurationSeconds)).
		Scan(ctx, &results)
	if err != nil || len(results) == 0 {
		return 0, err
	}
	return results[0].Sum, nil
}
