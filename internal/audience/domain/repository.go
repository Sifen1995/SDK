package domain

import (
	"context"
	"time"
)

// SegmentRepository loads Audiencemart segment definitions.
type SegmentRepository interface {
	GetByID(ctx context.Context, id string) (*AudienceSegment, error)
	GetByName(ctx context.Context, name string) (*AudienceSegment, error)
	Create(ctx context.Context, seg *AudienceSegment) error
	Update(ctx context.Context, seg *AudienceSegment) error
	// ListAvailableNow returns active segments within their availability window.
	ListAvailableNow(ctx context.Context, now time.Time) ([]AudienceSegment, error)
	ListAll(ctx context.Context) ([]AudienceSegment, error)
	// FindActiveByIntentSignal returns an active catalog segment whose signals include intentName.
	FindActiveByIntentSignal(ctx context.Context, intentName string, now time.Time) (*AudienceSegment, error)
}

// PurchaseRepository reads and writes segment purchase entitlements.
type PurchaseRepository interface {
	GetValidForCampaign(ctx context.Context, campaignID string, now time.Time) (*SegmentPurchase, error)
	CreatePurchase(ctx context.Context, purchase *SegmentPurchase) error
	// CreatePurchaseTx writes the purchase inside an outer unit-of-work transaction.
	CreatePurchaseTx(ctx context.Context, tx any, purchase *SegmentPurchase) error
}
