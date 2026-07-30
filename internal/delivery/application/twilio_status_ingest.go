package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"

	deliverydomain "skykin-platform/internal/delivery/domain"
)

type TwilioStatusIngestService struct {
	attempts  deliverydomain.SMSSendAttemptRepository
	authToken string
}

func NewTwilioStatusIngestService(
	attempts deliverydomain.SMSSendAttemptRepository,
	authToken string,
) *TwilioStatusIngestService {
	return &TwilioStatusIngestService{
		attempts:  attempts,
		authToken: strings.TrimSpace(authToken),
	}
}

func (s *TwilioStatusIngestService) VerifySignature(
	rawURL string,
	values url.Values,
	signature string,
) bool {
	if strings.TrimSpace(s.authToken) == "" || strings.TrimSpace(signature) == "" {
		return false
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	payload := rawURL
	for _, key := range keys {
		for _, value := range values[key] {
			payload += key + value
		}
	}
	mac := hmac.New(sha1.New, []byte(s.authToken))
	mac.Write([]byte(payload))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *TwilioStatusIngestService) ProcessStatus(
	ctx context.Context,
	sendKey, messageSID, messageStatus string,
) error {
	if s == nil || s.attempts == nil {
		return fmt.Errorf("twilio status ingest is not configured")
	}
	attempt, err := s.attempts.FindBySendKey(ctx, strings.TrimSpace(sendKey))
	if err != nil {
		return err
	}
	status := mapTwilioStatus(messageStatus)
	return s.attempts.UpdateStatus(ctx, attempt.ID, status, strings.TrimSpace(messageSID), "")
}

func mapTwilioStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "queued", "accepted", "sending", "sent":
		return deliverydomain.SMSSendStatusSent
	case "delivered":
		return deliverydomain.SMSSendStatusDelivered
	case "undelivered", "failed":
		return deliverydomain.SMSSendStatusFailed
	case "clicked":
		return deliverydomain.SMSSendStatusClicked
	default:
		return deliverydomain.SMSSendStatusSent
	}
}
