package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skykin-platform/internal/analytics/domain"
)

// AggregateQueue enqueues anonymized aggregate batches for async persistence.
type AggregateQueue interface {
	Enqueue(ctx context.Context, report *domain.AggregateReport) error
}

// AggregateIngestService accepts SDK aggregate batches and hands them to a queue.
type AggregateIngestService struct {
	queue AggregateQueue
}

func NewAggregateIngestService(queue AggregateQueue) *AggregateIngestService {
	return &AggregateIngestService{queue: queue}
}

// EnqueueReport validates the batch and pushes it to queue:analytics_aggregate.
func (s *AggregateIngestService) EnqueueReport(ctx context.Context, report *domain.AggregateReport) error {
	if s == nil || s.queue == nil {
		return fmt.Errorf("aggregate ingest is not configured")
	}
	if err := ValidateAggregateReport(report); err != nil {
		return err
	}
	normalizeAggregateReport(report)
	return s.queue.Enqueue(ctx, report)
}

// ValidateAggregateReport checks the anonymous device batch contract.
func ValidateAggregateReport(report *domain.AggregateReport) error {
	if report == nil {
		return fmt.Errorf("aggregate report is required")
	}
	if len(report.Intents) == 0 {
		return fmt.Errorf("intents must not be empty")
	}
	for i, item := range report.Intents {
		if strings.TrimSpace(item.IntentName) == "" {
			return fmt.Errorf("intents[%d].intent_name is required", i)
		}
		if item.Count < 1 {
			return fmt.Errorf("intents[%d].count must be at least 1", i)
		}
		if item.DaysConsistent < 1 {
			return fmt.Errorf("intents[%d].days_consistent must be at least 1", i)
		}
	}
	return nil
}

func normalizeAggregateReport(report *domain.AggregateReport) {
	if report.DateBucket.IsZero() {
		now := time.Now().UTC()
		report.DateBucket = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		b := report.DateBucket.UTC()
		report.DateBucket = time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	}
	for i := range report.Intents {
		report.Intents[i].IntentName = strings.TrimSpace(report.Intents[i].IntentName)
	}
}
