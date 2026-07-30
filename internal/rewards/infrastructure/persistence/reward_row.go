package persistence

import (
	"time"

	"skykin-platform/internal/rewards/domain"
)

// RewardRow is the GORM persistence model for rewards.
type RewardRow struct {
	ID             string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PseudonymousID string `gorm:"column:pseudonymous_id;type:uuid;not null;index"`
	IntentID       string `gorm:"type:uuid;not null"`
	RuleID     string     `gorm:"type:uuid;not null"`
	RewardType string     `gorm:"type:varchar(50);not null"`
	Amount     float64    `gorm:"type:numeric(10,2);not null"`
	Currency   string     `gorm:"type:varchar(50);not null"`
	Status     string     `gorm:"type:varchar(20);default:'pending'"`
	Message    string     `gorm:"type:text;not null"`
	CreatedAt  time.Time  `gorm:"not null;default:now()"`
	SentAt     *time.Time
	ClaimedAt  *time.Time
}

func (RewardRow) TableName() string { return "rewards" }

func (row *RewardRow) ToDomain() *domain.Reward {
	if row == nil {
		return nil
	}
	return &domain.Reward{
		ID:             row.ID,
		PseudonymousID: row.PseudonymousID,
		IntentID:       row.IntentID,
		RuleID:     row.RuleID,
		RewardType: row.RewardType,
		Amount:     row.Amount,
		Currency:   row.Currency,
		Status:     row.Status,
		Message:    row.Message,
		CreatedAt:  row.CreatedAt,
		SentAt:     row.SentAt,
		ClaimedAt:  row.ClaimedAt,
	}
}

func RewardRowFromDomain(r *domain.Reward) *RewardRow {
	if r == nil {
		return nil
	}
	return &RewardRow{
		ID:             r.ID,
		PseudonymousID: r.PseudonymousID,
		IntentID:       r.IntentID,
		RuleID:     r.RuleID,
		RewardType: r.RewardType,
		Amount:     r.Amount,
		Currency:   r.Currency,
		Status:     r.Status,
		Message:    r.Message,
		CreatedAt:  r.CreatedAt,
		SentAt:     r.SentAt,
		ClaimedAt:  r.ClaimedAt,
	}
}

// RewardRuleRow is the GORM persistence model for reward_rules.
type RewardRuleRow struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IntentName string    `gorm:"type:varchar(100);unique;not null"`
	RewardType string    `gorm:"type:varchar(50);not null"`
	Amount     float64   `gorm:"type:numeric(10,2);not null"`
	Currency   string    `gorm:"type:varchar(50);not null"`
	Message    string    `gorm:"type:text;not null"`
	IsActive   bool      `gorm:"default:true"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
}

func (RewardRuleRow) TableName() string { return "reward_rules" }

func (row *RewardRuleRow) ToDomain() *domain.RewardRule {
	if row == nil {
		return nil
	}
	return &domain.RewardRule{
		ID:         row.ID,
		IntentName: row.IntentName,
		RewardType: row.RewardType,
		Amount:     row.Amount,
		Currency:   row.Currency,
		Message:    row.Message,
		IsActive:   row.IsActive,
		CreatedAt:  row.CreatedAt,
	}
}

func RewardRuleRowFromDomain(r *domain.RewardRule) *RewardRuleRow {
	if r == nil {
		return nil
	}
	return &RewardRuleRow{
		ID:         r.ID,
		IntentName: r.IntentName,
		RewardType: r.RewardType,
		Amount:     r.Amount,
		Currency:   r.Currency,
		Message:    r.Message,
		IsActive:   r.IsActive,
		CreatedAt:  r.CreatedAt,
	}
}
