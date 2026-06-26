package domain

import (
	"context"
	"time"

	"skykin-platform/internal/audience/model"
)

// SegmentRepository loads Audiencemart segment definitions.
type SegmentRepository interface {
	GetByID(ctx context.Context, id string) (*model.AudienceSegment, error)
	GetByName(ctx context.Context, name string) (*model.AudienceSegment, error)
	Create(ctx context.Context, seg *model.AudienceSegment) error
	// ListAvailableNow returns active segments within their availability window.
	ListAvailableNow(ctx context.Context, now time.Time) ([]model.AudienceSegment, error)
}

// PurchaseRepository reads segment purchase entitlements.
type PurchaseRepository interface {
	// GetValidForCampaign returns the purchase row when still within valid_from / valid_until.
	GetValidForCampaign(ctx context.Context, campaignID string, now time.Time) (*model.SegmentPurchase, error)
}
