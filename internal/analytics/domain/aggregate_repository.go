package domain

import (
	"context"
	"time"
)

// AggregateUpsertItem is one intent counter delta applied to a date bucket.
type AggregateUpsertItem struct {
	IntentName     string
	Count          int
	DaysConsistent float64
}

// AggregateRepository upserts daily anonymized intent aggregate counters.
type AggregateRepository interface {
	UpsertBatch(ctx context.Context, dateBucket time.Time, items []AggregateUpsertItem) error
}
