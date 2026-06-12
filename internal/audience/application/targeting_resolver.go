package application

import (
	"context"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/campaigns/model"

	"github.com/google/uuid"
)

// IntentUserFinder resolves user IDs from stored intent predictions.
type IntentUserFinder interface {
	FindUsersWithIntent(ctx context.Context, intentName string, minConfidence float64, since time.Time) ([]uuid.UUID, error)
	FindUsersWithAnyIntent(ctx context.Context, intentNames []string, minConfidence float64, since time.Time) ([]uuid.UUID, error)
}

// TargetingResolver applies segment purchase rules when matching users to campaigns.
type TargetingResolver struct {
	segments  audiencedomain.SegmentRepository
	purchases audiencedomain.PurchaseRepository
}

func NewTargetingResolver(segments audiencedomain.SegmentRepository, purchases audiencedomain.PurchaseRepository) *TargetingResolver {
	return &TargetingResolver{segments: segments, purchases: purchases}
}

// ResolveUserIDs returns candidate user IDs for a campaign, respecting segment entitlements.
func (r *TargetingResolver) ResolveUserIDs(
	ctx context.Context,
	campaign *model.Campaign,
	intents IntentUserFinder,
	minConfidence float64,
	since time.Time,
) ([]uuid.UUID, error) {
	if campaign.SegmentID == nil || *campaign.SegmentID == "" {
		return intents.FindUsersWithIntent(ctx, campaign.TargetIntent, minConfidence, since)
	}

	now := time.Now().UTC()
	if _, err := r.purchases.GetValidForCampaign(ctx, campaign.ID, now); err != nil {
		// No valid purchase — segment-targeted campaign delivers to nobody.
		return nil, nil
	}

	seg, err := r.segments.GetByID(ctx, *campaign.SegmentID)
	if err != nil {
		return nil, err
	}
	signals, err := ParseIntentSignals(seg)
	if err != nil {
		return nil, err
	}
	// Users must match the campaign target_intent and belong to the segment signal pool.
	if !TargetIntentAllowed(signals, campaign.TargetIntent) {
		return nil, nil
	}
	return intents.FindUsersWithIntent(ctx, campaign.TargetIntent, minConfidence, since)
}
