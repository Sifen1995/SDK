package infrastructure

import (
	"context"
	"fmt"

	deliverydomain "skykin-platform/internal/delivery/domain"
	"skykin-platform/internal/delivery/infrastructure/persistence"

	"gorm.io/gorm"
)

// DeliveryLogRepository writes campaign_delivery_logs from the delivery bounded context.
type DeliveryLogRepository struct {
	db *gorm.DB
}

func NewDeliveryLogRepository(db *gorm.DB) *DeliveryLogRepository {
	return &DeliveryLogRepository{db: db}
}

var _ deliverydomain.DeliveryLogRepository = (*DeliveryLogRepository)(nil)

func (r *DeliveryLogRepository) CreateBatch(ctx context.Context, logs []deliverydomain.DeliveryLog) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("delivery log repository is not configured")
	}
	if len(logs) == 0 {
		return nil
	}
	rows := make([]persistence.DeliveryLogRow, 0, len(logs))
	for i := range logs {
		row := persistence.DeliveryLogRowFromDomain(&logs[i])
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
