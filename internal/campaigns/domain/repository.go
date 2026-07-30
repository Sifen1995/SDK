package domain

import "context"

// CampaignRepository loads and mutates campaigns for targeting and portal use cases.
type CampaignRepository interface {
	ListActive(ctx context.Context) ([]Campaign, error)
	ListPendingModeration(ctx context.Context) ([]Campaign, error)
	Create(ctx context.Context, c *Campaign) error
	CreateTx(ctx context.Context, tx any, c *Campaign) error
	Transaction(ctx context.Context, fn func(tx any) error) error
	Get(ctx context.Context, id string) (*Campaign, error)
	ListByAdvertiser(ctx context.Context, advertiserID string) ([]Campaign, error)
	ListAll(ctx context.Context) ([]Campaign, error)
	Update(ctx context.Context, c *Campaign) error
	FindActiveForIntent(ctx context.Context, targetIntent, channelCode string) (*Campaign, error)
	ListEligibleForDelivery(ctx context.Context, intentName, channelCode string) ([]Campaign, error)
	ListActiveMaster(ctx context.Context) ([]Campaign, error)
	ListActiveByIntent(ctx context.Context, intentName string) ([]Campaign, error)
	CountActiveByAdvertiser(ctx context.Context, advertiserID string) (int, error)
}
