package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"
	billingmodel "skykin-platform/internal/billing/model"

	"gorm.io/gorm"
)

// defaultPurchaseDays is the MVP entitlement window for a segment on one campaign.
const defaultPurchaseDays = 30

// impressionBundle is the fixed impression count used to price segment purchases (MVP).
const impressionBundle = 1000

// PurchaseQuote holds validated segment pricing before the campaign row exists.
type PurchaseQuote struct {
	AdvertiserID string
	SegmentID    string
	AmountPaid   float64
	ValidFrom    time.Time
	ValidUntil   time.Time
}

// PurchaseService validates segments and records purchases in segment_purchases.
type PurchaseService struct {
	segments audiencedomain.SegmentRepository
}

func NewPurchaseService(segments audiencedomain.SegmentRepository) *PurchaseService {
	return &PurchaseService{segments: segments}
}

// PreparePurchase validates the segment is available and the plan allows Audiencemart.
func (s *PurchaseService) PreparePurchase(ctx context.Context, advertiserID, segmentID string, plan billingmodel.SubscriptionPlan) (*PurchaseQuote, error) {
	if !plan.AudiencemartEnabled {
		return nil, fmt.Errorf("plan %q does not include Audiencemart", plan.Name)
	}
	seg, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		return nil, errors.New("audience segment not found")
	}
	return s.prepareFromSegment(advertiserID, segmentID, seg)
}

// ValidateTargetIntent checks that the campaign intent is included in the segment definition.
func (s *PurchaseService) ValidateTargetIntent(ctx context.Context, segmentID, targetIntent string) error {
	seg, err := s.segments.GetByID(ctx, segmentID)
	if err != nil {
		return errors.New("audience segment not found")
	}
	signals, err := ParseIntentSignals(seg)
	if err != nil {
		return err
	}
	if !TargetIntentAllowed(signals, targetIntent) {
		return fmt.Errorf("target_intent %q is not in segment %q signals", targetIntent, seg.Name)
	}
	return nil
}

func (s *PurchaseService) prepareFromSegment(
	advertiserID, segmentID string,
	seg *model.AudienceSegment,
) (*PurchaseQuote, error) {
	now := time.Now().UTC()
	if !seg.IsActive {
		return nil, errors.New("audience segment is not active")
	}
	if now.Before(seg.AvailableFrom) {
		return nil, errors.New("audience segment is not yet available")
	}
	if seg.AvailableUntil != nil && now.After(*seg.AvailableUntil) {
		return nil, errors.New("audience segment has expired")
	}

	// MVP pricing: estimated_cpm × fixed impression bundle / 1000
	amount := seg.EstimatedCPM * float64(impressionBundle) / 1000.0
	validUntil := now.AddDate(0, 0, defaultPurchaseDays)

	return &PurchaseQuote{
		AdvertiserID: advertiserID,
		SegmentID:    segmentID,
		AmountPaid:   amount,
		ValidFrom:    now,
		ValidUntil:   validUntil,
	}, nil
}

// ConfirmPurchaseTx writes the entitlement row inside the campaign create transaction.
func (s *PurchaseService) ConfirmPurchaseTx(ctx context.Context, tx *gorm.DB, quote *PurchaseQuote, campaignID string) error {
	if quote == nil {
		return nil
	}
	purchase := &model.SegmentPurchase{
		AdvertiserID: quote.AdvertiserID,
		SegmentID:    quote.SegmentID,
		CampaignID:   campaignID,
		AmountPaid:   quote.AmountPaid,
		ValidFrom:    quote.ValidFrom,
		ValidUntil:   quote.ValidUntil,
	}
	return tx.WithContext(ctx).Create(purchase).Error
}
