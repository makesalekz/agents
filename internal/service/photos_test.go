package service

import (
	"context"
	"testing"

	v1 "gitlab.calendaria.team/services/agents/api/agents/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Story 6.3: Photo Reports ---

func TestAddVisitPhoto(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	checkin, _ := svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: route.Route.Points[0].Id,
		Lat:          "43.238",
		Lon:          "76.945",
	})

	resp, err := svc.AddVisitPhoto(ctx, &v1.AddVisitPhotoRequest{
		VisitId:  checkin.Visit.Id,
		MediaUrl: "https://media.example.com/photo1.jpg",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Photo)
	assert.Equal(t, checkin.Visit.Id, resp.Photo.VisitId)
	assert.Equal(t, "https://media.example.com/photo1.jpg", resp.Photo.MediaUrl)
	assert.NotEmpty(t, resp.Photo.CreatedAt)
}

func TestAddVisitPhoto_NoTenant(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := context.Background()

	_, err := svc.AddVisitPhoto(ctx, &v1.AddVisitPhotoRequest{
		VisitId:  1,
		MediaUrl: "https://media.example.com/photo1.jpg",
	})
	require.Error(t, err)
}

func TestAddVisitPhoto_EmptyURL(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.AddVisitPhoto(ctx, &v1.AddVisitPhotoRequest{
		VisitId:  1,
		MediaUrl: "",
	})
	require.Error(t, err)
}

func TestGetVisitPhotos(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	route, _ := svc.CreateRoute(ctx, &v1.CreateRouteRequest{
		AgentId: 10,
		Date:    "2026-06-01",
		Points:  []*v1.RoutePointInput{{StoreTenantId: 100, OrderNum: 1}},
	})
	checkin, _ := svc.CheckIn(ctx, &v1.CheckInRequest{
		RoutePointId: route.Route.Points[0].Id,
		Lat:          "43.238",
		Lon:          "76.945",
	})

	svc.AddVisitPhoto(ctx, &v1.AddVisitPhotoRequest{
		VisitId:  checkin.Visit.Id,
		MediaUrl: "https://media.example.com/photo1.jpg",
	})
	svc.AddVisitPhoto(ctx, &v1.AddVisitPhotoRequest{
		VisitId:  checkin.Visit.Id,
		MediaUrl: "https://media.example.com/photo2.jpg",
	})

	resp, err := svc.GetVisitPhotos(ctx, &v1.GetVisitPhotosRequest{
		VisitId: checkin.Visit.Id,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 2)
}

func TestGetVisitPhotos_Empty(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	resp, err := svc.GetVisitPhotos(ctx, &v1.GetVisitPhotosRequest{
		VisitId: 999,
	})
	require.NoError(t, err)
	assert.Len(t, resp.Items, 0)
}
