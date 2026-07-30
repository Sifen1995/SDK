package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNoValidPurchase means the campaign targets a segment it is no longer entitled to.
var ErrNoValidPurchase = errors.New("no valid segment purchase for campaign")

// IntentUserFinder resolves pseudonymous ids from stored intent predictions.
type IntentUserFinder interface {
	FindPseudonymousIDsWithIntent(ctx context.Context, intentName string, minConfidence float64, since time.Time) ([]string, error)
	FindPseudonymousIDsWithAnyIntent(ctx context.Context, intentNames []string, minConfidence float64, since time.Time) ([]string, error)
}

// SegmentTarget is the narrow campaign projection audience targeting needs.
type SegmentTarget struct {
	CampaignID   string
	SegmentID    string
	TargetIntent string
}

// TargetingResolver applies segment purchase rules when matching audiences to campaigns.
type TargetingResolver struct {
	segments   audiencedomain.SegmentRepository
	purchases  audiencedomain.PurchaseRepository
	membership audiencedomain.MembershipRepository
}

func NewTargetingResolver(
	segments audiencedomain.SegmentRepository,
	purchases audiencedomain.PurchaseRepository,
	membership audiencedomain.MembershipRepository,
) *TargetingResolver {
	return &TargetingResolver{segments: segments, purchases: purchases, membership: membership}
}

// Resolve returns the pseudonymous ids a campaign may be delivered to.
//
// Campaigns without a segment fall back to live intent matching. Segment-targeted
// campaigns deliver only to the purchased segment's membership, and only while the
// purchase is inside its valid_from/valid_until window.
func (r *TargetingResolver) Resolve(
	ctx context.Context,
	target SegmentTarget,
	intents IntentUserFinder,
	minConfidence float64,
	since time.Time,
) ([]string, error) {
	if target.SegmentID == "" {
		return intents.FindPseudonymousIDsWithIntent(ctx, target.TargetIntent, minConfidence, since)
	}

	now := time.Now().UTC()
	if _, err := r.purchases.GetValidForCampaign(ctx, target.CampaignID, now); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoValidPurchase
		}
		return nil, fmt.Errorf("load segment purchase: %w", err)
	}

	seg, err := r.segments.GetByID(ctx, target.SegmentID)
	if err != nil {
		return nil, fmt.Errorf("load segment: %w", err)
	}
	signals, err := ParseIntentSignals(seg)
	if err != nil {
		return nil, err
	}
	if !TargetIntentAllowed(signals, target.TargetIntent) {
		return nil, nil
	}

	segID, err := uuid.Parse(target.SegmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid segment id %q: %w", target.SegmentID, err)
	}
	return r.membership.FindPseudonymousIDsInSegment(ctx, segID)
}
