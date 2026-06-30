package application

import (
	"context"

	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"
	campaignEvents "skykin-platform/internal/campaigns/events"
)

// RecordSegmentPurchaseUseCase records a segment purchase after campaign creation.
type RecordSegmentPurchaseUseCase struct {
	purchases audiencedomain.PurchaseRepository
}

func NewRecordSegmentPurchaseUseCase(purchases audiencedomain.PurchaseRepository) *RecordSegmentPurchaseUseCase {
	return &RecordSegmentPurchaseUseCase{purchases: purchases}
}

func (uc *RecordSegmentPurchaseUseCase) Execute(ctx context.Context, evt campaignEvents.CampaignCreatedEvent) error {
	if !evt.HasPurchase {
		return nil
	}
	return uc.purchases.CreatePurchase(ctx, &model.SegmentPurchase{
		AdvertiserID: evt.AdvertiserID,
		SegmentID:    evt.SegmentID,
		CampaignID:   evt.CampaignID,
		AmountPaid:   evt.AmountPaid,
		ValidFrom:    evt.ValidFrom,
		ValidUntil:   evt.ValidUntil,
	})
}
