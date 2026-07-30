package infrastructure

import (
	"context"
	"log/slog"
	"time"

	"skykin-platform/internal/analytics/domain"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// StartAnalyticsAggregateWorker drains queue:analytics_aggregate and upserts intent_aggregate_counts.
func StartAnalyticsAggregateWorker(db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	if db == nil || rdb == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	repo := NewAggregateRepository(db)
	go runAnalyticsAggregateWorker(context.Background(), rdb, repo, logger)
}

func runAnalyticsAggregateWorker(
	ctx context.Context,
	rdb *platformredis.RedisClient,
	repo domain.AggregateRepository,
	logger *slog.Logger,
) {
	for {
		if ctx.Err() != nil {
			return
		}

		msg, err := rdb.BRPop(ctx, AnalyticsAggregateQueueKey, 2*time.Second)
		if err != nil || msg == "" {
			continue
		}

		report, err := DecodeAggregateQueuePayload(msg)
		if err != nil {
			logger.Warn("analytics aggregate worker: invalid payload", "error", err)
			continue
		}

		if err := applyAggregateReport(ctx, repo, report); err != nil {
			logger.Error("analytics aggregate worker: upsert failed",
				"date_bucket", report.DateBucket.Format("2006-01-02"),
				"intents", len(report.Intents),
				"error", err,
			)
			continue
		}
		logger.Info("analytics aggregate worker: applied batch",
			"date_bucket", report.DateBucket.Format("2006-01-02"),
			"intents", len(report.Intents),
		)
	}
}

func applyAggregateReport(
	ctx context.Context,
	repo domain.AggregateRepository,
	report *domain.AggregateReport,
) error {
	return repo.UpsertBatch(ctx, report.DateBucket, toAggregateItems(report))
}

func toAggregateItems(report *domain.AggregateReport) []domain.AggregateUpsertItem {
	items := make([]domain.AggregateUpsertItem, len(report.Intents))
	for i, item := range report.Intents {
		items[i] = domain.AggregateUpsertItem{
			IntentName:     item.IntentName,
			Count:          item.Count,
			DaysConsistent: item.DaysConsistent,
		}
	}
	return items
}
