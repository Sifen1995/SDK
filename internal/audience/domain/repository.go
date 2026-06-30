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
	Update(ctx context.Context, seg *model.AudienceSegment) error
	// ListAvailableNow returns active segments within their availability window.
	ListAvailableNow(ctx context.Context, now time.Time) ([]model.AudienceSegment, error)
	ListAll(ctx context.Context) ([]model.AudienceSegment, error)
}

// PurchaseRepository reads and writes segment purchase entitlements.
type PurchaseRepository interface {
	GetValidForCampaign(ctx context.Context, campaignID string, now time.Time) (*model.SegmentPurchase, error)
	CreatePurchase(ctx context.Context, purchase *model.SegmentPurchase) error
}
