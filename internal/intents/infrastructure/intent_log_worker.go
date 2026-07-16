package infrastructure

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	intentdomain "skykin-platform/internal/intents/domain"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

const intentLogBatchSize = 50
const intentLogFlushInterval = 3 * time.Second

// StartIntentLogWorker drains queue:intent_logs and bulk-inserts into Postgres.
func StartIntentLogWorker(db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	if db == nil || rdb == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	repo := NewIntentRepository(db, nil)
	go runIntentLogWorker(context.Background(), rdb, repo, logger)
}

func runIntentLogWorker(
	ctx context.Context,
	rdb *platformredis.RedisClient,
	repo intentdomain.IntentRepository,
	logger *slog.Logger,
) {
	batch := make([]*intentdomain.Intent, 0, intentLogBatchSize)
	lastFlush := time.Now()

	for {
		if ctx.Err() != nil {
			flushIntentBatch(ctx, repo, batch, logger)
			return
		}

		msg, err := rdb.BRPop(ctx, IntentLogQueueKey, 2*time.Second)
		if err == nil && msg != "" {
			var entry IntentLogEntry
			if err := json.Unmarshal([]byte(msg), &entry); err != nil {
				logger.Warn("intent log worker: invalid payload", "error", err)
			} else {
				batch = append(batch, entry.toIntent())
			}
		}

		shouldFlush := len(batch) >= intentLogBatchSize ||
			(len(batch) > 0 && time.Since(lastFlush) >= intentLogFlushInterval)
		if shouldFlush {
			flushIntentBatch(ctx, repo, batch, logger)
			batch = batch[:0]
			lastFlush = time.Now()
		}
	}
}

func flushIntentBatch(
	ctx context.Context,
	repo intentdomain.IntentRepository,
	batch []*intentdomain.Intent,
	logger *slog.Logger,
) {
	if len(batch) == 0 {
		return
	}
	if err := repo.CreateBatch(ctx, batch); err != nil {
		logger.Error("intent log worker: batch insert failed", "count", len(batch), "error", err)
		return
	}
	logger.Info("intent log worker: flushed batch", "count", len(batch))
}
