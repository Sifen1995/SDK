package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	deliverydomain "skykin-platform/internal/delivery/domain"
)

// StreamDeliveryFields is the transport-neutral projection of a billing stream message.
type StreamDeliveryFields struct {
	CampaignID     string
	EventType      string
	PseudonymousID string
	Source         string
	InstallToken   string
	OccurredAt     string
}

// MapStreamToDeliveryLog converts a stream entry into a delivery-owned log row.
func MapStreamToDeliveryLog(fields StreamDeliveryFields, clickTokenSecret string) (*deliverydomain.DeliveryLog, error) {
	campaignID := strings.TrimSpace(fields.CampaignID)
	eventType := strings.ToLower(strings.TrimSpace(fields.EventType))
	if campaignID == "" || eventType == "" {
		return nil, fmt.Errorf("campaign_id and event_type are required")
	}

	status := statusForEvent(eventType)
	if status == "" {
		return nil, fmt.Errorf("unsupported event_type for delivery log")
	}
	if eventType == "install" {
		if err := validateInstallToken(fields.InstallToken, campaignID, clickTokenSecret); err != nil {
			return nil, err
		}
	}

	pseudonymousID := strings.TrimSpace(fields.PseudonymousID)
	sessionID := "telemetry"
	if pseudonymousID == "" || strings.TrimSpace(fields.Source) == "anonymous" {
		pseudonymousID = deliverydomain.AnonymousPseudonymousID
		sessionID = "anonymous"
	}

	return &deliverydomain.DeliveryLog{
		CampaignID:     campaignID,
		PseudonymousID: pseudonymousID,
		SessionID:      sessionID,
		DeliveryStatus: status,
		LoggedAt:       parseOccurredAt(fields.OccurredAt),
	}, nil
}

func statusForEvent(eventType string) string {
	switch eventType {
	case "impression":
		return deliverydomain.StatusRendered
	case "click":
		return deliverydomain.StatusClicked
	case "install", "signup", "purchase":
		return deliverydomain.StatusConverted
	default:
		return ""
	}
}

func validateInstallToken(token, campaignID, secret string) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("invalid install token")
	}
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(campaignID))
	expected := hex.EncodeToString(h.Sum(nil))
	if !hmac.Equal([]byte(token), []byte(expected)) {
		return fmt.Errorf("invalid install token")
	}
	return nil
}

func parseOccurredAt(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC()
		}
	}
	return time.Now().UTC()
}
