package infrastructure

import (
	"context"
	"fmt"

	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/billing/infrastructure/persistence"

	"gorm.io/gorm"
)

// BillingEventRepository persists billable campaign interactions.
type BillingEventRepository struct {
	db *gorm.DB
}

func NewBillingEventRepository(db *gorm.DB) *BillingEventRepository {
	return &BillingEventRepository{db: db}
}

// CreateBatch inserts many billing events in one multi-row statement.
func (r *BillingEventRepository) CreateBatch(ctx context.Context, events []billingdomain.BillingEvent) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("billing event repository is not configured")
	}
	if len(events) == 0 {
		return nil
	}
	rows := make([]persistence.BillingEventRow, 0, len(events))
	for i := range events {
		row := persistence.BillingEventRowFromDomain(&events[i])
		if row == nil {
			continue
		}
		rows = append(rows, *row)
	}
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(rows, 100).Error
}
