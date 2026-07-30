package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	analyticsdomain "skykin-platform/internal/analytics/domain"
	"skykin-platform/internal/analytics/infrastructure/persistence"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AggregateRepository upserts daily intent aggregate counters for analytics.
type AggregateRepository struct {
	db *gorm.DB
}

func NewAggregateRepository(db *gorm.DB) *AggregateRepository {
	return &AggregateRepository{db: db}
}

var _ analyticsdomain.AggregateRepository = (*AggregateRepository)(nil)

// UpsertBatch applies ON CONFLICT (intent_name, date_bucket) for each intent item.
func (r *AggregateRepository) UpsertBatch(
	ctx context.Context,
	dateBucket time.Time,
	items []analyticsdomain.AggregateUpsertItem,
) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("aggregate repository is not configured")
	}
	if len(items) == 0 {
		return nil
	}

	bucket := dateBucket
	if bucket.IsZero() {
		bucket = time.Now().UTC()
	}
	bucket = time.Date(bucket.Year(), bucket.Month(), bucket.Day(), 0, 0, 0, 0, time.UTC)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			intentName := strings.TrimSpace(item.IntentName)
			if intentName == "" || item.Count < 1 {
				continue
			}
			row := persistence.IntentAggregateCountRow{
				IntentName:    intentName,
				DateBucket:    bucket,
				SignalCount:   item.Count,
				WeightedCount: item.DaysConsistent,
			}
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "intent_name"}, {Name: "date_bucket"}},
				DoUpdates: clause.Assignments(map[string]any{
					"signal_count":   gorm.Expr("intent_aggregate_counts.signal_count + EXCLUDED.signal_count"),
					"weighted_count": gorm.Expr("intent_aggregate_counts.weighted_count + EXCLUDED.weighted_count"),
				}),
			}).Create(&row).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
