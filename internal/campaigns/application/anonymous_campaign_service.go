package application

import (
	"context"
	"fmt"

	campaigndomain "skykin-platform/internal/campaigns/domain"
)

// ActiveMasterCampaignReader loads the anonymous delivery master campaign list.
type ActiveMasterCampaignReader interface {
	ListActiveMasterCampaigns(ctx context.Context) ([]campaigndomain.Campaign, error)
}

// AnonymousCampaignService serves the non-consented campaign master list.
type AnonymousCampaignService struct {
	campaigns ActiveMasterCampaignReader
}

func NewAnonymousCampaignService(campaigns ActiveMasterCampaignReader) *AnonymousCampaignService {
	return &AnonymousCampaignService{campaigns: campaigns}
}

// ListActiveMaster returns all active, un-exhausted campaigns (no frequency filtering).
func (s *AnonymousCampaignService) ListActiveMaster(ctx context.Context) ([]campaigndomain.Campaign, error) {
	if s == nil || s.campaigns == nil {
		return nil, fmt.Errorf("anonymous campaign service is not configured")
	}
	return s.campaigns.ListActiveMasterCampaigns(ctx)
}
