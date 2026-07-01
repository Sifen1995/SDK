package domain

import "context"

// CampaignRepository loads campaigns for targeting and portal use cases.
type CampaignRepository interface {
	ListActive(ctx context.Context) ([]Campaign, error)
}
