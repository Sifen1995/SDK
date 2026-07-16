package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"skykin-platform/internal/analytics/domain"
	platformredis "skykin-platform/internal/platform/redis"
)

const AnalyticsAggregateQueueKey = "queue:analytics_aggregate"

// AggregateQueuePayload is the Redis list JSON for anonymous aggregate batches.
type AggregateQueuePayload struct {
	DateBucket string                    `json:"date_bucket"`
	Intents    []AggregateQueueIntentItem `json:"intents"`
}

// AggregateQueueIntentItem is one intent entry in the queued batch.
type AggregateQueueIntentItem struct {
	IntentName     string  `json:"intent_name"`
	Count          int     `json:"count"`
	DaysConsistent float64 `json:"days_consistent"`
}

// AnalyticsAggregateQueue pushes aggregate batches onto Redis for a background worker.
type AnalyticsAggregateQueue struct {
	rdb *platformredis.RedisClient
}

func NewAnalyticsAggregateQueue(rdb *platformredis.RedisClient) *AnalyticsAggregateQueue {
	return &AnalyticsAggregateQueue{rdb: rdb}
}

func (q *AnalyticsAggregateQueue) Enqueue(ctx context.Context, report *domain.AggregateReport) error {
	if q == nil || q.rdb == nil {
		return fmt.Errorf("analytics aggregate queue is not configured")
	}
	if report == nil {
		return fmt.Errorf("aggregate report is required")
	}

	payload := AggregateQueuePayload{
		DateBucket: report.DateBucket.UTC().Format("2006-01-02"),
		Intents:    make([]AggregateQueueIntentItem, len(report.Intents)),
	}
	for i, item := range report.Intents {
		payload.Intents[i] = AggregateQueueIntentItem{
			IntentName:     item.IntentName,
			Count:          item.Count,
			DaysConsistent: item.DaysConsistent,
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return q.rdb.RPush(ctx, AnalyticsAggregateQueueKey, string(raw))
}

func DecodeAggregateQueuePayload(raw string) (*domain.AggregateReport, error) {
	var payload AggregateQueuePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	bucket, err := time.Parse("2006-01-02", payload.DateBucket)
	if err != nil {
		return nil, fmt.Errorf("invalid date_bucket: %w", err)
	}
	report := &domain.AggregateReport{
		DateBucket: bucket.UTC(),
		Intents:    make([]domain.AggregateIntentSignal, len(payload.Intents)),
	}
	for i, item := range payload.Intents {
		report.Intents[i] = domain.AggregateIntentSignal{
			IntentName:     item.IntentName,
			Count:          item.Count,
			DaysConsistent: item.DaysConsistent,
		}
	}
	return report, nil
}
