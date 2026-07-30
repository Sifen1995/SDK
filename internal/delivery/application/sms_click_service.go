package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	deliverydomain "skykin-platform/internal/delivery/domain"
	"skykin-platform/internal/platform/redis"
)

type SMSClickService struct {
	attempts deliverydomain.SMSSendAttemptRepository
	rdb      *redis.RedisClient
}

func NewSMSClickService(
	attempts deliverydomain.SMSSendAttemptRepository,
	rdb *redis.RedisClient,
) *SMSClickService {
	return &SMSClickService{attempts: attempts, rdb: rdb}
}

func (s *SMSClickService) ProcessClick(ctx context.Context, token string) (string, error) {
	if s == nil || s.attempts == nil || s.rdb == nil {
		return "", fmt.Errorf("sms click service is not configured")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token is required")
	}
	attempt, err := s.attempts.FindByTrackingToken(ctx, token)
	if err != nil {
		return "", err
	}
	if attempt.Status != deliverydomain.SMSSendStatusClicked {
		if _, err := s.rdb.XAdd(ctx, BillingEventsStream, billingEventsStreamMax, map[string]interface{}{
			"campaign_id":       attempt.CampaignID,
			"event_type":        "click",
			"transaction_value": "0.0000",
			"occurred_at":       time.Now().UTC().Format(time.RFC3339Nano),
			"pseudonymous_id":   attempt.PseudonymousID,
		}); err != nil {
			return "", err
		}
		if err := s.attempts.UpdateStatus(ctx, attempt.ID, deliverydomain.SMSSendStatusClicked, attempt.ProviderMessageID, ""); err != nil {
			return "", err
		}
	}
	return attempt.DestinationURL, nil
}
