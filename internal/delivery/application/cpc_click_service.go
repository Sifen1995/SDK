package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"skykin-platform/internal/platform/redis"
)

type CPCClickService struct {
	secretKey []byte
	rdb       *redis.RedisClient
}


func NewCPCClickService(secretKey string, rdb *redis.RedisClient) *CPCClickService {
	return &CPCClickService{
		secretKey: []byte(secretKey),
		rdb:       rdb,
	}
}

func (s *CPCClickService) ProcessClick(ctx context.Context, campaignID, token string) error {
	if !s.validateToken(campaignID, token) {
		return fmt.Errorf("invalid token signature or expired timestamp")
	}

	_, err := s.rdb.XAdd(ctx, BillingEventsStream, billingEventsStreamMax, map[string]interface{}{
		"campaign_id":       campaignID,
		"event_type":        "click",
		"transaction_value": "0.0000",
		"occurred_at":       time.Now().UTC().Format(time.RFC3339Nano),
		"source":            "anonymous",
	})
	return err
}

func (s *CPCClickService) validateToken(campaignID, token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	sig := strings.TrimSpace(parts[0])
	hourBucket := strings.TrimSpace(parts[1])
	if sig == "" || hourBucket == "" {
		return false
	}

	now := time.Now().UTC()
	currentBucket := now.Format("2006-01-02-15")
	prevBucket := now.Add(-1 * time.Hour).Format("2006-01-02-15")

	if hourBucket != currentBucket && hourBucket != prevBucket {
		return false
	}

	message := fmt.Sprintf("%s:%s", campaignID, hourBucket)
	h := hmac.New(sha256.New, s.secretKey)
	h.Write([]byte(message))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedSig))
}
