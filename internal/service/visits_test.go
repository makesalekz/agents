package service

import (
	"context"
	"testing"

	v1 "github.com/makesalekz/agents/api/agents/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Story 6.2: GPS Visits ---

func TestCheckIn(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	// Create a route with a point
	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	pointID := route.Route.Points[0].Id

	resp, err := svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: pointID,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Visit)
	assert.Equal(t, pointID, resp.Visit.RoutePointId)
	assert.Equal(t, int64(10), resp.Visit.AgentId)
	assert.Equal(t, "43.238", resp.Visit.CheckinLat)
	assert.Equal(t, "76.945", resp.Visit.CheckinLon)
	assert.NotEmpty(t, resp.Visit.CheckinAt)
	assert.Empty(t, resp.Visit.CheckoutAt) // not checked out yet
}

func TestCheckIn_NoTenant(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := context.Background()

	_, err := svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: 1,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.Error(t, err)
}

func TestCheckIn_InvalidRoutePoint(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: 999,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.Error(t, err)
}

func TestCheckIn_DuplicateCheckin(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	pointID := route.Route.Points[0].Id

	_, err := svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: pointID,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.NoError(t, err)

	// Second checkin should fail
	_, err = svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: pointID,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.Error(t, err)
}

func TestCheckIn_TenantIsolation(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx1 := ctxWithTenantAndActor(1, 10)
	ctx2 := ctxWithTenantAndActor(2, 20)

	route, _ := svc.CreateRoute(ctx1, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	pointID := route.Route.Points[0].Id

	// Tenant 2 should not be able to check in to tenant 1's route point
	_, err := svc.CheckIn(ctx2, &v1.CheckInRequest{
		RoutePointId: pointID,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.Error(t, err)
}

func TestCheckOut(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	pointID := route.Route.Points[0].Id

	// Check in first
	_, _ = svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: pointID,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})

	// Check out
	resp, err := svc.CheckOut(ctx, &v1.CheckOutRequest{
		RoutePointId: pointID,
		Lat:          "43.2385",
		Lon:          "76.9455",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Visit)
	assert.NotEmpty(t, resp.Visit.CheckoutAt)
	assert.Equal(t, "43.2385", resp.Visit.CheckoutLat)
	assert.Equal(t, "76.9455", resp.Visit.CheckoutLon)
	assert.True(t, resp.Visit.DurationSeconds >= 0)
}

func TestCheckOut_WithoutCheckin(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.CheckOut(ctx, &v1.CheckOutRequest{
		RoutePointId: 999,
		Lat:          "43.2380",
		Lon:          "76.9450",
	})
	require.Error(t, err)
}

func TestCheckOut_DuplicateCheckout(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	pointID := route.Route.Points[0].Id

	svc.CheckIn(ctx, &v1.CheckInRequest{RoutePointId: pointID, Lat: "43.238", Lon: "76.945"})
	svc.CheckOut(ctx, &v1.CheckOutRequest{RoutePointId: pointID, Lat: "43.238", Lon: "76.945"})

	// Second checkout should fail
	_, err := svc.CheckOut(ctx, &v1.CheckOutRequest{
		RoutePointId: pointID,
		Lat:          "43.238",
		Lon:          "76.945",
	})
	require.Error(t, err)
}

func TestListVisits(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points: []*v1.RoutePointInput{
			{StoreTenantId: 100, OrderNum: 1},
			{StoreTenantId: 200, OrderNum: 2},
		},
	})

	svc.CheckIn(ctx, &v1.CheckInRequest{RoutePointId: route.Route.Points[0].Id, Lat: "43.238", Lon: "76.945"})
	svc.CheckIn(ctx, &v1.CheckInRequest{RoutePointId: route.Route.Points[1].Id, Lat: "43.240", Lon: "76.950"})

	resp, err := svc.ListVisits(ctx, &v1.ListVisitsRequest{AgentId: 10})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 2)
	assert.NotNil(t, resp.Paginate)
}

func TestListVisits_TenantIsolation(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx1 := ctxWithTenantAndActor(1, 10)
	ctx2 := ctxWithTenantAndActor(2, 20)

	route, _ := svc.CreateRoute(ctx1, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	svc.CheckIn(ctx1, &v1.CheckInRequest{RoutePointId: route.Route.Points[0].Id, Lat: "43.238", Lon: "76.945"})

	resp, err := svc.ListVisits(ctx2, &v1.ListVisitsRequest{AgentId: 10})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 0)
}
