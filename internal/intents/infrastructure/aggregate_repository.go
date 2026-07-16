package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	intentdomain "skykin-platform/internal/intents/domain"
	"skykin-platform/internal/intents/infrastructure/persistence"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AggregateUpsertItem is one intent counter delta applied to a date bucket.
type AggregateUpsertItem struct {
	IntentName     string
	Count          int
	DaysConsistent float64
}

// AggregateRepository upserts daily intent aggregate counters.
type AggregateRepository struct {
	db *gorm.DB
}

func NewAggregateRepository(db *gorm.DB) *AggregateRepository {
	return &AggregateRepository{db: db}
}

// UpsertBatch applies ON CONFLICT (intent_name, date_bucket) for each intent item.
func (r *AggregateRepository) UpsertBatch(
	ctx context.Context,
	dateBucket time.Time,
	items []AggregateUpsertItem,
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
			// Use columns (not ON CONSTRAINT): GORM uniqueIndex creates a unique INDEX,
			// which Postgres rejects for ON CONFLICT ON CONSTRAINT.
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

// Ingest increments signal_count by Count (default 1) and weighted_count by days_consistent.
func (r *AggregateRepository) Ingest(
	ctx context.Context,
	signal *intentdomain.IntentAggregateSignal,
) (*intentdomain.IntentAggregateCount, error) {
	if signal == nil {
		return nil, fmt.Errorf("aggregate signal is required")
	}
	count := signal.Count
	if count < 1 {
		count = 1
	}
	if err := r.UpsertBatch(ctx, signal.DateBucket, []AggregateUpsertItem{{
		IntentName:     signal.IntentName,
		Count:          count,
		DaysConsistent: signal.DaysConsistent,
	}}); err != nil {
		return nil, err
	}

	bucket := signal.DateBucket
	if bucket.IsZero() {
		bucket = time.Now().UTC()
	}
	bucket = time.Date(bucket.Year(), bucket.Month(), bucket.Day(), 0, 0, 0, 0, time.UTC)

	var out persistence.IntentAggregateCountRow
	if err := r.db.WithContext(ctx).
		Where("intent_name = ? AND date_bucket = ?", strings.TrimSpace(signal.IntentName), bucket).
		First(&out).Error; err != nil {
		return nil, err
	}
	return out.ToDomain(), nil
}
