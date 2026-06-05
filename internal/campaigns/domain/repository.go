package domain

import (
	"context"

	"skykin-platform/internal/campaigns/model"
)

// CampaignRepository loads campaigns for targeting and portal use cases.
type CampaignRepository interface {
	ListActive(ctx context.Context) ([]model.Campaign, error)
}
