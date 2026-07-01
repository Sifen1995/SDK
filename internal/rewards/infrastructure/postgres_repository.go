package infrastructure

import (
	"context"
	"fmt"
	"sync"

	rewardsdomain "skykin-platform/internal/rewards/domain"
	"skykin-platform/internal/rewards/infrastructure/persistence"

	"gorm.io/gorm"
)

type postgresRewardRepository struct {
	db    *gorm.DB
	rules map[string]rewardsdomain.RewardRule
	mu    sync.RWMutex
}

func NewRewardRepository(db *gorm.DB) rewardsdomain.RewardRepository {
	repo := &postgresRewardRepository{
		db:    db,
		rules: make(map[string]rewardsdomain.RewardRule),
	}
	_ = repo.RefreshRules(context.Background())
	return repo
}

func (r *postgresRewardRepository) RefreshRules(ctx context.Context) error {
	var rows []persistence.RewardRuleRow
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&rows).Error; err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = make(map[string]rewardsdomain.RewardRule)
	for _, row := range rows {
		rule := *row.ToDomain()
		r.rules[rule.IntentName] = rule
	}
	return nil
}

func (r *postgresRewardRepository) GetRuleByIntent(ctx context.Context, intent string) (*rewardsdomain.RewardRule, error) {
	r.mu.RLock()
	rule, exists := r.rules[intent]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no active reward rule found for intent: %s", intent)
	}
	return &rule, nil
}

func (r *postgresRewardRepository) CreateReward(ctx context.Context, reward *rewardsdomain.Reward) error {
	row := persistence.RewardRowFromDomain(reward)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*reward = *row.ToDomain()
	return nil
}
