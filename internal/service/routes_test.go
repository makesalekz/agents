package service

import (
	"context"
	"testing"

	v1 "gitlab.calendaria.team/services/agents/api/agents/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Story 6.1: Routes ---

func TestCreateRoute(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	resp, err := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points: []*v1.RoutePointInput{
			{StoreTenantId: 100, OrderNum: 1},
			{StoreTenantId: 200, OrderNum: 2},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Route)
	assert.Equal(t, int64(10), resp.Route.AgentId)
	assert.Equal(t, "2026-06-01", resp.Route.Date)
	assert.Equal(t, "PLANNED", resp.Route.Status)
	assert.Equal(t, int64(1), resp.Route.TenantId)
	assert.Len(t, resp.Route.Points, 2)
	assert.Equal(t, int64(100), resp.Route.Points[0].StoreTenantId)
	assert.Equal(t, int32(1), resp.Route.Points[0].OrderNum)
	assert.Equal(t, int64(200), resp.Route.Points[1].StoreTenantId)
	assert.Equal(t, int32(2), resp.Route.Points[1].OrderNum)
}

func TestCreateRoute_NoTenant(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := context.Background()

	_, err := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	require.Error(t, err)
}

func TestCreateRoute_InvalidDate(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	_, err := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "invalid",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	require.Error(t, err)
}

func TestCreateRoute_NoPoints(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	_, err := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
	})
	require.Error(t, err)
}

func TestCreateRoute_NoAgentID(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	_, err := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		Date:   "2026-06-01",
		Points: []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	require.Error(t, err)
}

func TestGetRoute(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	created, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})

	resp, err := svc.GetRoute(ctx, &v1.GetRouteRequest{Id: created.Route.Id})
	require.NoError(t, err)
	assert.Equal(t, created.Route.Id, resp.Route.Id)
	assert.Equal(t, int64(10), resp.Route.AgentId)
}

func TestGetRoute_NotFound(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	_, err := svc.GetRoute(ctx, &v1.GetRouteRequest{Id: 999})
	require.Error(t, err)
}

func TestGetRoute_TenantIsolation(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx1 := ctxWithTenant(1)
	ctx2 := ctxWithTenant(2)

	created, _ := svc.CreateRoute(ctx1, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})

	_, err := svc.GetRoute(ctx2, &v1.GetRouteRequest{Id: created.Route.Id})
	require.Error(t, err)
}

func TestListRoutes(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-02",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 200, OrderNum: 1}},
	})

	resp, err := svc.ListRoutes(ctx, &v1.ListRoutesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 2)
	assert.NotNil(t, resp.Paginate)
}

func TestListRoutes_FilterByAgent(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenant(1)

	svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 20,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 200, OrderNum: 1}},
	})

	resp, err := svc.ListRoutes(ctx, &v1.ListRoutesRequest{AgentId: 10})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
	assert.Equal(t, int64(10), resp.Items[0].AgentId)
}

func TestListRoutes_TenantIsolation(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx1 := ctxWithTenant(1)
	ctx2 := ctxWithTenant(2)

	svc.CreateRoute(ctx1, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	svc.CreateRoute(ctx2, &v1.CreateRouteRequest{
		AgentId: 20,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 200, OrderNum: 1}},
	})

	resp, err := svc.ListRoutes(ctx1, &v1.ListRoutesRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 1)
}
