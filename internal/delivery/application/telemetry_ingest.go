package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	BillingEventsStream    = "stream:billing_events"
	billingEventsStreamMax = 100000

	telemetryDedupKeyPrefix = "lock:telemetry:"
	telemetryClickTTL       = time.Hour
	telemetryImpressionTTL  = 5 * time.Minute
)

// TelemetryPublisher is the Redis surface needed for telemetry ingest.
type TelemetryPublisher interface {
	SetNX(ctx context.Context, key string, value string, ttl time.Duration) (bool, error)
	XAdd(ctx context.Context, stream string, maxLen int64, values map[string]interface{}) (string, error)
}

// ConsentedTrackCommand is a consented SDK telemetry event.
type ConsentedTrackCommand struct {
	CampaignID       string
	EventType        string
	PseudonymousID   string
	TransactionValue float64
	OccurredAt       string
	InstallToken     string
}

// AnonymousTrackCommand is a non-consented impression track.
type AnonymousTrackCommand struct {
	CampaignID string
	EventType  string
}

// TelemetryIngestService owns validation, dedup, and stream enqueue rules.
type TelemetryIngestService struct {
	publisher TelemetryPublisher
}

func NewTelemetryIngestService(publisher TelemetryPublisher) *TelemetryIngestService {
	return &TelemetryIngestService{publisher: publisher}
}

func (s *TelemetryIngestService) TrackConsented(ctx context.Context, cmd ConsentedTrackCommand) error {
	if s == nil || s.publisher == nil {
		return fmt.Errorf("telemetry stream unavailable")
	}

	eventType := strings.ToLower(strings.TrimSpace(cmd.EventType))
	switch eventType {
	case "impression", "click", "install", "signup", "purchase":
	default:
		return fmt.Errorf("invalid event_type: must be impression, click, install, signup, or purchase")
	}

	campaignID := strings.TrimSpace(cmd.CampaignID)
	pseudonymousID := strings.TrimSpace(cmd.PseudonymousID)
	if eventType == "install" && strings.TrimSpace(cmd.InstallToken) == "" {
		return fmt.Errorf("install_token is required for install events")
	}

	if ttl, ok := telemetryDedupTTL(eventType); ok {
		if pseudonymousID == "" {
			return fmt.Errorf("pseudonymous_id required for impression and click deduplication")
		}
		lockKey := telemetryDedupKeyPrefix + pseudonymousID + ":" + campaignID + ":" + eventType
		acquired, err := s.publisher.SetNX(ctx, lockKey, "1", ttl)
		if err != nil {
			return fmt.Errorf("telemetry dedup failed: %w", err)
		}
		if !acquired {
			return nil
		}
	}

	occurredAt := strings.TrimSpace(cmd.OccurredAt)
	if occurredAt == "" {
		occurredAt = time.Now().UTC().Format(time.RFC3339Nano)
	} else if _, err := time.Parse(time.RFC3339Nano, occurredAt); err != nil {
		if _, err2 := time.Parse(time.RFC3339, occurredAt); err2 != nil {
			return fmt.Errorf("invalid occurred_at: expected RFC3339")
		}
	}

	values := map[string]interface{}{
		"campaign_id":       campaignID,
		"event_type":        eventType,
		"transaction_value": strconv.FormatFloat(cmd.TransactionValue, 'f', 4, 64),
		"occurred_at":       occurredAt,
	}
	if eventType == "install" {
		values["install_token"] = strings.TrimSpace(cmd.InstallToken)
	}
	if pseudonymousID != "" {
		values["pseudonymous_id"] = pseudonymousID
	}
	if _, err := s.publisher.XAdd(ctx, BillingEventsStream, billingEventsStreamMax, values); err != nil {
		return fmt.Errorf("telemetry enqueue failed: %w", err)
	}
	return nil
}

func (s *TelemetryIngestService) TrackAnonymous(ctx context.Context, cmd AnonymousTrackCommand) error {
	if s == nil || s.publisher == nil {
		return fmt.Errorf("telemetry stream unavailable")
	}
	eventType := strings.ToLower(strings.TrimSpace(cmd.EventType))
	if eventType != "impression" {
		return fmt.Errorf("anonymous track currently accepts impression only")
	}
	_, err := s.publisher.XAdd(ctx, BillingEventsStream, billingEventsStreamMax, map[string]interface{}{
		"campaign_id":       strings.TrimSpace(cmd.CampaignID),
		"event_type":        eventType,
		"transaction_value": "0.0000",
		"occurred_at":       time.Now().UTC().Format(time.RFC3339Nano),
		"source":            "anonymous",
	})
	if err != nil {
		return fmt.Errorf("telemetry enqueue failed: %w", err)
	}
	return nil
}

func telemetryDedupTTL(eventType string) (time.Duration, bool) {
	switch eventType {
	case "click":
		return telemetryClickTTL, true
	case "impression":
		return telemetryImpressionTTL, true
	default:
		return 0, false
	}
}
