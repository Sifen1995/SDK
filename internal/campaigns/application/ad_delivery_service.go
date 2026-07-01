package application

import (
	"context"
	"fmt"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/campaigns/infrastructure"
)

// AdDeliveryService builds WebSocket ad payloads for matched intents.
type AdDeliveryService struct {
	repo *infrastructure.Repository
}

func NewAdDeliveryService(repo *infrastructure.Repository) *AdDeliveryService {
	return &AdDeliveryService{repo: repo}
}

type AdPayload struct {
	Type         string         `json:"type"`
	Intent       string         `json:"intent"`
	CampaignID   string         `json:"campaign_id"`
	CampaignName string         `json:"campaign_name"`
	ChannelCode  string         `json:"channel_code"`
	Content      map[string]any `json:"content"`
}

// BuildAdForIntent finds an active campaign for the intent across supported channels.
func (s *AdDeliveryService) BuildAdForIntent(ctx context.Context, intent string) (*AdPayload, error) {
	for _, code := range []string{"IN_APP_BANNER", "SMS_PLUS", "PUSH", "NATIVE_FEED"} {
		c, err := s.repo.FindActiveForIntent(ctx, intent, code)
		if err != nil {
			continue
		}
		content, err := infrastructure.CampaignAdContent(c, code)
		if err != nil {
			continue
		}
		return &AdPayload{
			Type:         "campaign_ad",
			Intent:       intent,
			CampaignID:   c.ID,
			CampaignName: c.Name,
			ChannelCode:  code,
			Content:      content,
		}, nil
	}
	return nil, fmt.Errorf("no active campaign for intent %s", intent)
}

func (s *AdDeliveryService) LogDispatched(ctx context.Context, campaignID, userID, sessionID string) error {
	return s.repo.LogDelivery(ctx, &campaigndomain.DeliveryLog{
		CampaignID:     campaignID,
		UserID:         userID,
		SessionID:      sessionID,
		DeliveryStatus: campaigndomain.DeliveryDispatched,
	})
}
