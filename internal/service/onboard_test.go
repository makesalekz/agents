package service

import (
	"context"
	"testing"

	v1 "gitlab.calendaria.team/services/agents/api/agents/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Story 6.4: New Store Onboarding ---

func TestOnboardStore(t *testing.T) {
	svc, _, _, _, onboardingsRepo := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	resp, err := svc.OnboardStore(ctx, &v1.OnboardStoreRequest{
		StoreName:  "Магазин Арман",
		OwnerPhone: "+77001234567",
	})
	require.NoError(t, err)
	assert.True(t, resp.NewTenantId > 0)

	// Verify onboarding was recorded
	assert.Equal(t, 1, len(onboardingsRepo.onboardings))
	for _, o := range onboardingsRepo.onboardings {
		assert.Equal(t, int64(1), o.TenantID)
		assert.Equal(t, int64(10), o.AgentID)
		assert.Equal(t, "Магазин Арман", o.StoreName)
		assert.Equal(t, "+77001234567", o.OwnerPhone)
		assert.Equal(t, resp.NewTenantId, o.NewTenantID)
	}
}

func TestOnboardStore_NoTenant(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := context.Background()

	_, err := svc.OnboardStore(ctx, &v1.OnboardStoreRequest{
		StoreName:  "Магазин",
		OwnerPhone: "+77001234567",
	})
	require.Error(t, err)
}

func TestOnboardStore_EmptyStoreName(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.OnboardStore(ctx, &v1.OnboardStoreRequest{
		StoreName:  "",
		OwnerPhone: "+77001234567",
	})
	require.Error(t, err)
}

func TestOnboardStore_EmptyPhone(t *testing.T) {
	svc, _, _, _, _ := setupService()
	ctx := ctxWithTenantAndActor(1, 10)

	_, err := svc.OnboardStore(ctx, &v1.OnboardStoreRequest{
		StoreName:  "Магазин",
		OwnerPhone: "",
	})
	require.Error(t, err)
}
