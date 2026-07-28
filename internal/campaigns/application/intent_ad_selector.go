package application

import (
	"context"
	"fmt"

	"skykin-platform/internal/campaigns/infrastructure"
	intentsApp "skykin-platform/internal/intents/application"
)

// IntentAdSelector implements intents/application.AdSelector using the cached plan-tier ranker.
type IntentAdSelector struct {
	campaigns   *infrastructure.CachedCampaignRepository
	linkBuilder *infrastructure.PlayLinkBuilder
}

var _ intentsApp.AdSelector = (*IntentAdSelector)(nil)

func NewIntentAdSelector(campaigns *infrastructure.CachedCampaignRepository, linkBuilder *infrastructure.PlayLinkBuilder) *IntentAdSelector {
	return &IntentAdSelector{campaigns: campaigns, linkBuilder: linkBuilder}
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
		content, err := infrastructure.CampaignAdContent(campaign, code, s.linkBuilder)
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
	return best, nil
}
