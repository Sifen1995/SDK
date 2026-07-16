package application_test

import (
	"testing"
	"time"

	analyticsApp "skykin-platform/internal/analytics/application"
	"skykin-platform/internal/analytics/domain"
)

func TestValidateAggregateReport_OK(t *testing.T) {
	err := analyticsApp.ValidateAggregateReport(&domain.AggregateReport{
		DateBucket: time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC),
		Intents: []domain.AggregateIntentSignal{
			{IntentName: "fashion_interest", Count: 3, DaysConsistent: 7},
		},
	})
	if err != nil {
		t.Fatalf("expected valid report, got %v", err)
	}
}

func TestValidateAggregateReport_EmptyIntents(t *testing.T) {
	err := analyticsApp.ValidateAggregateReport(&domain.AggregateReport{
		DateBucket: time.Now().UTC(),
		Intents:    nil,
	})
	if err == nil {
		t.Fatal("expected error for empty intents")
	}
}

func TestValidateAggregateReport_BadCount(t *testing.T) {
	err := analyticsApp.ValidateAggregateReport(&domain.AggregateReport{
		Intents: []domain.AggregateIntentSignal{
			{IntentName: "fashion_interest", Count: 0, DaysConsistent: 7},
		},
	})
	if err == nil {
		t.Fatal("expected error for count < 1")
	}
}
