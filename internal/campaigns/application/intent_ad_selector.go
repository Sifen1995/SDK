package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/campaigns/infrastructure"
	deliverydomain "skykin-platform/internal/delivery/domain"
	intentsApp "skykin-platform/internal/intents/application"
)

// IntentAdSelector implements intents/application.AdSelector using the cached plan-tier ranker.
type IntentAdSelector struct {
	campaigns    *infrastructure.CachedCampaignRepository
	campaignRepo *infrastructure.Repository
	deliveryJobs deliverydomain.DeliveryRepository
	log          *slog.Logger
}

var _ intentsApp.AdSelector = (*IntentAdSelector)(nil)

func NewIntentAdSelector(
	campaigns *infrastructure.CachedCampaignRepository,
	campaignRepo *infrastructure.Repository,
	deliveryJobs deliverydomain.DeliveryRepository,
	log *slog.Logger,
) *IntentAdSelector {
	if log == nil {
		log = slog.Default()
	}
	return &IntentAdSelector{
		campaigns:    campaigns,
		campaignRepo: campaignRepo,
		deliveryJobs: deliveryJobs,
		log:          log,
	}
}

// SelectAd finds the highest-plan-tier eligible campaign for an intent and channel.
func (s *IntentAdSelector) SelectAd(
	ctx context.Context,
	pseudonymousID, targetIntent, channelCode string,
) (*intentsApp.AdSelection, error) {
	if s == nil || s.campaigns == nil {
		return nil, fmt.Errorf("ad selector is not configured")
	}

	codes := []string{channelCode}
	if channelCode == "" {
		codes = []string{"IN_APP_BANNER", "SMS_PLUS", "PUSH", "NATIVE_FEED"}
	}

	var best *intentsApp.AdSelection
	var bestPlanFee float64
	for _, code := range codes {
		if code == "" {
			continue
		}
		campaign, err := s.campaigns.SelectBestCampaign(ctx, targetIntent, code, pseudonymousID)
		if err != nil {
			continue
		}
		content, err := infrastructure.CampaignAdContent(campaign, code)
		if err != nil {
			continue
		}
		if best == nil || campaign.PlanMonthlyFeeETB > bestPlanFee {
			bestPlanFee = campaign.PlanMonthlyFeeETB
			best = &intentsApp.AdSelection{
				CampaignID:   campaign.ID,
				CampaignName: campaign.Name,
				ChannelCode:  code,
				Content:      content,
			}
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no active campaign for intent %s", targetIntent)
	}

	s.recordDispatch(ctx, pseudonymousID, best.CampaignID)
	return best, nil
}

func (s *IntentAdSelector) recordDispatch(ctx context.Context, userID, campaignID string) {
	if userID == "" || campaignID == "" {
		return
	}
	if s.campaignRepo != nil {
		if err := s.campaignRepo.LogDelivery(ctx, &campaigndomain.DeliveryLog{
			CampaignID:     campaignID,
			UserID:         userID,
			SessionID:      "ingest-ad",
			DeliveryStatus: campaigndomain.DeliveryDispatched,
			LoggedAt:       time.Now().UTC(),
		}); err != nil {
			s.log.Warn("intent ad selector: delivery log failed", "campaign_id", campaignID, "error", err)
		}
	}
	if s.deliveryJobs != nil {
		if err := s.deliveryJobs.RecordJob(ctx, userID, campaignID); err != nil {
			s.log.Warn("intent ad selector: delivery_jobs failed", "campaign_id", campaignID, "error", err)
		}
	}
}
