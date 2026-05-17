package data

import (
	"context"
	"time"

	"gitlab.calendaria.team/services/agents/ent"
	"gitlab.calendaria.team/services/agents/ent/onboarding"
)

type OnboardingsRepo interface {
	Create(ctx context.Context, dto OnboardingDto) (*ent.Onboarding, error)
	CountByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) (int32, error)
}

type onboardingsRepo struct {
	db *ent.Client
}

func NewOnboardingsRepo(d *Data) OnboardingsRepo {
	return &onboardingsRepo{db: d.db}
}

func (r *onboardingsRepo) Create(ctx context.Context, dto OnboardingDto) (*ent.Onboarding, error) {
	return r.db.Onboarding.Create().
		SetTenantID(dto.TenantID).
		SetAgentID(dto.AgentID).
		SetNewTenantID(dto.NewTenantID).
		SetStoreName(dto.StoreName).
		SetOwnerPhone(dto.OwnerPhone).
		Save(ctx)
}

func (r *onboardingsRepo) CountByAgentAndDateRange(ctx context.Context, tenantID, agentID int64, from, to time.Time) (int32, error) {
	count, err := r.db.Onboarding.Query().
		Where(
			onboarding.TenantID(tenantID),
			onboarding.AgentID(agentID),
			onboarding.CreatedAtGTE(from),
			onboarding.CreatedAtLTE(to),
		).
		Count(ctx)
	return int32(count), err
}
