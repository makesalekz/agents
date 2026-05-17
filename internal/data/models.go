package data

import (
	"time"

	"github.com/shopspring/decimal"

	"gitlab.calendaria.team/services/agents/ent/enum"
)

type RouteDto struct {
	ID       int64
	TenantID int64
	AgentID  int64
	Date     time.Time
	Status   enum.RouteStatus
}

type RoutePointDto struct {
	RouteID       int64
	StoreTenantID int64
	OrderNum      int32
}

type VisitDto struct {
	TenantID     int64
	RoutePointID int64
	AgentID      int64
	CheckinLat   decimal.Decimal
	CheckinLon   decimal.Decimal
	CheckinAt    time.Time
}

type VisitPhotoDto struct {
	VisitID  int64
	MediaURL string
}

type OnboardingDto struct {
	TenantID   int64
	AgentID    int64
	NewTenantID int64
	StoreName  string
	OwnerPhone string
}
