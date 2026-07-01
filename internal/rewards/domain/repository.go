package domain

import "context"

// RewardRepository manages reward rules and issued rewards.
type RewardRepository interface {
	RefreshRules(ctx context.Context) error
	GetRuleByIntent(ctx context.Context, intent string) (*RewardRule, error)
	CreateReward(ctx context.Context, reward *Reward) error
}
