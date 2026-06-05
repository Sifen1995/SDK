package application

import (
	"context"
	"fmt"

	"skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/campaigns/model"
)

type AdDeliveryService struct {
	repo *infrastructure.Repository
}

func NewAdDeliveryService(repo *infrastructure.Repository) *AdDeliveryService {
	return &AdDeliveryService{repo: repo}
}

type AdPayload struct {
	Type           string         `json:"type"`
	Intent         string         `json:"intent"`
	CampaignID     string         `json:"campaign_id"`
	CampaignName   string         `json:"campaign_name"`
	CreativeFormat string         `json:"creative_format"`
	Content        map[string]any `json:"content"`
}

func (s *AdDeliveryService) BuildAdForIntent(ctx context.Context, intent string) (*AdPayload, error) {
	for _, format := range []string{"BANNER", "SMS_PLUS", "PUSH_PLUS"} {
		c, err := s.repo.FindActiveForIntent(ctx, intent, format)
		if err != nil {
			continue
		}
		content, err := infrastructure.CampaignAdContent(c)
		if err != nil {
			continue
		}
		return &AdPayload{
			Type:           "campaign_ad",
			Intent:         intent,
			CampaignID:     c.ID,
			CampaignName:   c.Name,
			CreativeFormat: c.CreativeFormat,
			Content:        content,
		}, nil
	}
	return nil, fmt.Errorf("no active campaign for intent %s", intent)
}

func (s *AdDeliveryService) LogDispatched(ctx context.Context, campaignID, userID, sessionID string) error {
	return s.repo.LogDelivery(ctx, &model.DeliveryLog{
		CampaignID:     campaignID,
		UserID:         userID,
		SessionID:      sessionID,
		DeliveryStatus: model.DeliveryDispatched,
	})
}
