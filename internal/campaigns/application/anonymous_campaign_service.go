package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
)

// ActiveMasterCampaignReader loads the anonymous delivery master campaign list.
type ActiveMasterCampaignReader interface {
	ListActiveMasterCampaigns(ctx context.Context) ([]campaigndomain.Campaign, error)
}

type CampaignWithClickToken struct {
	campaigndomain.Campaign
	ClickToken string `json:"click_token"`
}

// AnonymousCampaignService serves the non-consented campaign master list.
type AnonymousCampaignService struct {
	campaigns ActiveMasterCampaignReader
	secretKey []byte
}

func NewAnonymousCampaignService(campaigns ActiveMasterCampaignReader, secretKey string) *AnonymousCampaignService {
	return &AnonymousCampaignService{campaigns: campaigns, secretKey: []byte(secretKey)}
}

// ListActiveMaster returns all active, un-exhausted campaigns (no frequency filtering).
func (s *AnonymousCampaignService) ListActiveMaster(ctx context.Context) ([]CampaignWithClickToken, error) {
	if s == nil || s.campaigns == nil {
		return nil, fmt.Errorf("anonymous campaign service is not configured")
	}

	campaignList, err := s.campaigns.ListActiveMasterCampaigns(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	enrichedCampaigns := make([]CampaignWithClickToken, 0, len(campaignList))

	for _, campaign := range campaignList {
		token := s.GenerateClickToken(campaign.ID, now)
		enrichedCampaigns = append(enrichedCampaigns, CampaignWithClickToken{
			Campaign:   campaign,
			ClickToken: token,
		})
	}

	return enrichedCampaigns, nil
}

func (s *AnonymousCampaignService) GenerateClickToken(campaignID string, t time.Time) string {
	hourBucket := t.UTC().Format("2006-01-02-15") // e.g., "2026-07-27-13"
	message := fmt.Sprintf("%s:%s", campaignID, hourBucket)

	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(message))
	sig := hex.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s.%s", sig, hourBucket)
}
