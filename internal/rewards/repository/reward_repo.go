package repository

import (
	"context"
	"fmt"
	"skykin-platform/internal/rewards/model"
	"sync"

	"gorm.io/gorm"
)

type RewardRepository interface {
	GetRuleByIntent(ctx context.Context, intent string) (*model.RewardRule, error)
	CreateReward(ctx context.Context, reward *model.Reward) error
	RefreshRules(ctx context.Context) error
}

type rewardRepo struct {
	db    *gorm.DB
	rules map[string]model.RewardRule
	mu    sync.RWMutex
}

func NewRewardRepository(db *gorm.DB) RewardRepository {
	repo := &rewardRepo{
		db:    db,
		rules: make(map[string]model.RewardRule),
	}
	_ = repo.RefreshRules(context.Background())
	return repo
}

func (r *rewardRepo) RefreshRules(ctx context.Context) error {
	var activeRules []model.RewardRule
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&activeRules).Error; err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.rules = make(map[string]model.RewardRule)
	for _, rule := range activeRules {
		r.rules[rule.IntentName] = rule
	}
	return nil
}

func (r *rewardRepo) GetRuleByIntent(ctx context.Context, intent string) (*model.RewardRule, error) {
	r.mu.RLock()
	rule, exists := r.rules[intent]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no active reward rule found for intent: %s", intent)
	}

	return &rule, nil
}

func (r *rewardRepo) CreateReward(ctx context.Context, reward *model.Reward) error {
	return r.db.WithContext(ctx).Create(reward).Error
}
