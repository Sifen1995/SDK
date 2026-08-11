package domain

import (
	"context"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
)

type GeofenceZone struct {
	ID           string
	AdvertiserID string
	Latitude     float64
	Longitude    float64
	RadiusMetres int
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StoreVisit struct {
	ID             string
	PseudonymousID string
	ZoneID         string
	AccuracyM      *float64
	IsFlagged      bool
	FlagReason     string
	VisitedAt      time.Time
}

type CampaignGeofenceTarget struct {
	CampaignID string
	ZoneID     string
}

type ZoneRepository interface {
	Create(ctx context.Context, zone *GeofenceZone) error
	ListByAdvertiser(ctx context.Context, advertiserID string) ([]GeofenceZone, error)
	ListInactive(ctx context.Context) ([]GeofenceZone, error)
	GetByID(ctx context.Context, zoneID string) (*GeofenceZone, error)
	FindNearby(ctx context.Context, lat, lng float64, radiusMetres int) ([]GeofenceZone, error)
	// ActivateZone turns a single draft store live; no-op if already active.
	ActivateZone(ctx context.Context, zoneID string) (*GeofenceZone, error)
	// ActivateForCampaign activates inactive zones linked to the campaign only.
	ActivateForCampaign(ctx context.Context, campaignID string) error
}

type TargetRepository interface {
	Link(ctx context.Context, campaignID string, zoneIDs []string) error
	ListZonesForCampaign(ctx context.Context, campaignID string) ([]GeofenceZone, error)
	// ListEligibleCampaignsForZone returns live zone campaigns. When targetIntent is non-empty,
	// only campaigns with matching target_intent are returned.
	ListEligibleCampaignsForZone(ctx context.Context, zoneID, targetIntent string) ([]campaigndomain.Campaign, error)
}

type VisitRepository interface {
	Create(ctx context.Context, visit *StoreVisit) error
	// CountByUserExcluding counts store_visits for the user excluding one visit id (any zone).
	CountByUserExcluding(ctx context.Context, pseudonymousID, excludeVisitID string) (int64, error)
}

type LocationConsentRepository interface {
	HasLocationConsent(ctx context.Context, pseudonymousID string) (bool, error)
	SetLocationConsent(ctx context.Context, pseudonymousID string, consented bool) error
}

type FrequencyCapPort interface {
	IsFrequencyCapped(ctx context.Context, pseudonymousID, campaignID string, limit int) (bool, error)
	IncrementDeliveryCount(ctx context.Context, pseudonymousID, campaignID string, ttl time.Duration) (int64, error)
}

type BudgetExhaustedPort interface {
	IsBudgetExhausted(ctx context.Context, campaignID string) (bool, error)
}
