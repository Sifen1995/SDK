package application

import (
	"context"
	"fmt"

	intentsApp "skykin-platform/internal/intents/application"
	"skykin-platform/internal/campaigns/infrastructure"
)

// IntentAdSelector implements intents/application.AdSelector using campaigns infrastructure.
type IntentAdSelector struct {
	repo *infrastructure.Repository
}

var _ intentsApp.AdSelector = (*IntentAdSelector)(nil)

func NewIntentAdSelector(repo *infrastructure.Repository) *IntentAdSelector {
	return &IntentAdSelector{repo: repo}
}

// SelectAd finds the best active campaign for an intent and channel (campaigns module only).
func (s *IntentAdSelector) SelectAd(
	ctx context.Context,
	targetIntent, channelCode string,
) (*intentsApp.AdSelection, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("ad selector is not configured")
	}

	codes := []string{channelCode}
	if channelCode == "" {
		codes = []string{"IN_APP_BANNER", "SMS_PLUS", "PUSH", "NATIVE_FEED"}
	}

	for _, code := range codes {
		if code == "" {
			continue
		}
		campaign, err := s.repo.FindActiveForIntent(ctx, targetIntent, code)
		if err != nil {
			continue
		}
		content, err := infrastructure.CampaignAdContent(campaign, code)
		if err != nil {
			continue
		}
		return &intentsApp.AdSelection{
			CampaignID:   campaign.ID,
			CampaignName: campaign.Name,
			ChannelCode:  code,
			Content:      content,
		}, nil
	}

	return nil, fmt.Errorf("no active campaign for intent %s", targetIntent)
}
