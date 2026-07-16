package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	intentdomain "skykin-platform/internal/intents/domain"
	platformredis "skykin-platform/internal/platform/redis"
)

const IntentLogQueueKey = "queue:intent_logs"

// IntentLogEntry is the JSON payload written to the Redis intent log queue.
type IntentLogEntry struct {
	UserID       string    `json:"user_id"`
	IntentName   string    `json:"intent_name"`
	Confidence   float64   `json:"confidence"`
	ModelVersion string    `json:"model_version,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// IntentLogQueue enqueues intent profiles for asynchronous Postgres persistence.
type IntentLogQueue struct {
	rdb *platformredis.RedisClient
}

func NewIntentLogQueue(rdb *platformredis.RedisClient) *IntentLogQueue {
	return &IntentLogQueue{rdb: rdb}
}

func (q *IntentLogQueue) Enqueue(ctx context.Context, profile *intentdomain.IntentProfile) error {
	if q == nil || q.rdb == nil || profile == nil {
		return nil
	}
	entry := IntentLogEntry{
		UserID:       profile.PseudonymousID,
		IntentName:   profile.IntentName,
		Confidence:   profile.Confidence,
		ModelVersion: profile.ModelVersion,
		CreatedAt:    profile.RecordedAt,
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return q.rdb.RPush(ctx, IntentLogQueueKey, string(payload))
}

func (e IntentLogEntry) toIntent() *intentdomain.Intent {
	return &intentdomain.Intent{
		UserID:     e.UserID,
		IntentName: e.IntentName,
		Confidence: e.Confidence,
		CreatedAt:  e.CreatedAt,
	}
}
