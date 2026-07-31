package application

import (
	"context"
	"fmt"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	intentsApp "skykin-platform/internal/intents/application"
)

// EligibleCampaignSelector picks the best campaign for an intent/channel.
type EligibleCampaignSelector interface {
	SelectBestCampaign(ctx context.Context, intentName, channelCode, pseudonymousID string) (*campaigndomain.Campaign, error)
}

// IntentAdSelector implements intents/application.AdSelector using the cached plan-tier ranker.
type IntentAdSelector struct {
	campaigns EligibleCampaignSelector
}

var _ intentsApp.AdSelector = (*IntentAdSelector)(nil)

func NewIntentAdSelector(campaigns EligibleCampaignSelector) *IntentAdSelector {
	return &IntentAdSelector{campaigns: campaigns}
}

// SelectAd finds the highest-plan-tier eligible campaign for an intent and channel.
// When channelCode is empty and smsConsented is true, SMS_PLUS is tried first;
// a miss falls back to non-SMS channels. When smsConsented is false, SMS_PLUS is excluded.
func (s *IntentAdSelector) SelectAd(
	ctx context.Context,
	pseudonymousID, targetIntent, channelCode string,
	smsConsented bool,
) (*intentsApp.AdSelection, error) {
	if s == nil || s.campaigns == nil {
		return nil, fmt.Errorf("ad selector is not configured")
	}

	if channelCode != "" {
		if channelCode == "SMS_PLUS" && !smsConsented {
			return nil, fmt.Errorf("no active campaign for intent %s", targetIntent)
		}
		return s.selectOne(ctx, pseudonymousID, targetIntent, channelCode)
	}

	if smsConsented {
		if ad, err := s.selectOne(ctx, pseudonymousID, targetIntent, "SMS_PLUS"); err == nil {
			return ad, nil
		}
	}

	return s.selectBestAmong(ctx, pseudonymousID, targetIntent, []string{
		"IN_APP_BANNER", "PUSH", "NATIVE_FEED",
	})
}

func (s *IntentAdSelector) selectOne(
	ctx context.Context,
	pseudonymousID, targetIntent, channelCode string,
) (*intentsApp.AdSelection, error) {
	campaign, err := s.campaigns.SelectBestCampaign(ctx, targetIntent, channelCode, pseudonymousID)
	if err != nil {
		return nil, err
	}
	content, err := CampaignAdContent(campaign, channelCode)
	if err != nil {
		return nil, err
	}
	return &intentsApp.AdSelection{
		CampaignID:   campaign.ID,
		CampaignName: campaign.Name,
		ChannelCode:  channelCode,
		Content:      content,
		Campaign:     campaign,
	}, nil
}

func (s *IntentAdSelector) selectBestAmong(
	ctx context.Context,
	pseudonymousID, targetIntent string,
	codes []string,
) (*intentsApp.AdSelection, error) {
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
		content, err := CampaignAdContent(campaign, code)
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
				Campaign:     campaign,
			}
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no active campaign for intent %s", targetIntent)
	}
	return best, nil
}
