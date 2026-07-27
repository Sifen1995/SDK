package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"skykin-platform/internal/platform/redis"
)

type CPCClickService struct {
	secretKey []byte
	rdb       *redis.RedisClient
}

type AnonymousClickEvent struct {
	CampaignID string    `json:"campaign_id"`
	ClickedAt  time.Time `json:"clicked_at"`
	EventType  string    `json:"event_type"` // "CLICK"
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

	event := AnonymousClickEvent{
		CampaignID: campaignID,
		ClickedAt:  time.Now().UTC(),
		EventType:  "CLICK",
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.rdb.RPush(ctx, "queue:cpc_billing_events", string(data))
}

func (s *CPCClickService) validateToken(campaignID, token string) bool {
	var sig, hourBucket string
	if n, err := fmt.Sscanf(token, "%s.%s", &sig, &hourBucket); err != nil || n != 2 {
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
