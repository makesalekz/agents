package service

import (
	"context"
	"testing"
	"time"

	v1 "gitlab.calendaria.team/services/agents/api/agents/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Story 6.5: Agent Reports ---

func TestGetAgentReport(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	today := time.Now().Format("2006-01-02")

	// Create route with 2 points (date = today so visits match the range)
	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    today,
		Points: []*v1.RoutePointInput{
			{StoreTenantId: 100, OrderNum: 1},
			{StoreTenantId: 200, OrderNum: 2},
		},
	})

	// Check in to first point only
	svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: route.Route.Points[0].Id,
		Lat:          "43.238",
		Lon:          "76.945",
	})

	resp, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: today,
		DateTo:   today,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Report)
	assert.Equal(t, int64(10), resp.Report.AgentId)
	assert.Equal(t, int32(1), resp.Report.VisitsCount)
	assert.Equal(t, int32(0), resp.Report.OrdersCount)    // stub
	assert.Equal(t, int32(0), resp.Report.OnboardingsCount) // none yet
	assert.Equal(t, int32(2), resp.Report.TotalRoutePoints)
	assert.Equal(t, int32(1), resp.Report.CompletedRoutePoints)
	assert.Equal(t, float64(50), resp.Report.RouteCompletionPct)
}

func TestGetAgentReport_WithOnboarding(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	// Create onboarding
	svc.OnboardStore(ctx, &v1.OnboardStoreRequest{
		StoreName:  "Магазин",
		OwnerPhone: "+77001234567",
	})

	resp, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), resp.Report.OnboardingsCount)
}

func TestGetAgentReport_NoTenant(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := context.Background()

	_, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: "2026-06-01",
		DateTo:   "2026-06-30",
	})
	require.Error(t, err)
}

func TestGetAgentReport_InvalidDateFrom(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: "invalid",
		DateTo:   "2026-06-30",
	})
	require.Error(t, err)
}

func TestGetAgentReport_InvalidDateTo(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: "2026-06-01",
		DateTo:   "invalid",
	})
	require.Error(t, err)
}

func TestGetAgentReport_EmptyRange(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	resp, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: "2026-06-01",
		DateTo:   "2026-06-01",
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.Report.VisitsCount)
	assert.Equal(t, int32(0), resp.Report.TotalRoutePoints)
	assert.Equal(t, float64(0), resp.Report.RouteCompletionPct)
}

func TestGetAgentReport_FullCompletion(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	today := time.Now().Format("2006-01-02")

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    today,
		Points: []*v1.RoutePointInput{
			{StoreTenantId: 100, OrderNum: 1},
			{StoreTenantId: 200, OrderNum: 2},
		},
	})

	// Check in to both points
	svc.CheckIn(ctx, &v1.CheckInRequest{RoutePointId: route.Route.Points[0].Id, Lat: "43.238", Lon: "76.945"})
	svc.CheckIn(ctx, &v1.CheckInRequest{RoutePointId: route.Route.Points[1].Id, Lat: "43.240", Lon: "76.950"})

	resp, err := svc.GetAgentReport(ctx, &v1.GetAgentReportRequest{
		AgentId:  10,
		DateFrom: today,
		DateTo:   today,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(2), resp.Report.VisitsCount)
	assert.Equal(t, int32(2), resp.Report.CompletedRoutePoints)
	assert.Equal(t, float64(100), resp.Report.RouteCompletionPct)
}
