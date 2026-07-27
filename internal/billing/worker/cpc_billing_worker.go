package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"skykin-platform/internal/platform/redis"
)

type CPCWorker struct {
	rdb    *redis.RedisClient
	db     *gorm.DB
	logger *slog.Logger
}

type ClickQueuePayload struct {
	CampaignID string    `json:"campaign_id"`
	ClickedAt  time.Time `json:"clicked_at"`
	EventType  string    `json:"event_type"`
}

func NewCPCWorker(rdb *redis.RedisClient, db *gorm.DB, logger *slog.Logger) *CPCWorker {
	return &CPCWorker{rdb: rdb, db: db, logger: logger}
}

func (w *CPCWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := w.rdb.BRPop(ctx, "queue:cpc_billing_events", 2*time.Second)
			if err != nil || res == "" {
				continue
			}

			var payload ClickQueuePayload
			if err := json.Unmarshal([]byte(res), &payload); err != nil {
				w.logger.Error("failed to parse CPC click payload", "error", err)
				continue
			}

			if err := w.processCPCClick(ctx, payload); err != nil {
				w.logger.Error("failed to log CPC billing entry", "error", err)
			}
		}
	}
}

func (w *CPCWorker) processCPCClick(ctx context.Context, payload ClickQueuePayload) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Write interaction record to campaign_delivery_logs (user_id is NULL)
		sqlDelivery := `
			INSERT INTO campaign_delivery_logs (id, campaign_id, user_id, session_id, delivery_status, created_at)
			VALUES (?, ?, NULL, 'ANONYMOUS_CPC', 'CLICKED', ?);
		`
		if err := tx.Exec(sqlDelivery, uuid.New().String(), payload.CampaignID, payload.ClickedAt).Error; err != nil {
			return err
		}

		// 2. Write charge entry to billing_events (user_id is NULL)
		sqlBilling := `
			INSERT INTO billing_events (id, campaign_id, user_id, event_type, amount, channel_code, created_at)
			SELECT 
				?, 
				c.id, 
				NULL, 
				'CPC', 
				c.cpc_rate, 
				'IN_APP_BANNER', 
				?
			FROM campaigns c
			WHERE c.id = ?;
		`
		if err := tx.Exec(sqlBilling, uuid.New().String(), payload.ClickedAt, payload.CampaignID).Error; err != nil {
			return err
		}

		return nil
	})
}
