package worker

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	deliveryApp "skykin-platform/internal/delivery/application"
	deliverydomain "skykin-platform/internal/delivery/domain"
	deliveryInfra "skykin-platform/internal/delivery/infrastructure"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

const (
	billingEventsStream       = "stream:billing_events"
	deliveryLogProcessorGroup = "delivery_log_processor_group"
	deliveryLogConsumer       = "delivery-log-worker-1"
	deliveryReadBatchSize     = 100
	deliveryReadBlock         = 2 * time.Second
)

// StartDeliveryLogConsumer reads stream:billing_events with its own consumer group and
// writes only campaign_delivery_logs (no billing / campaign repository access).
func StartDeliveryLogConsumer(db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	if db == nil || rdb == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	if err := rdb.XGroupCreateMkStream(context.Background(), billingEventsStream, deliveryLogProcessorGroup, "0"); err != nil {
		logger.Error("delivery log consumer: create group failed", "error", err)
		return
	}
	go runDeliveryLogConsumer(context.Background(), db, rdb, logger)
}

func runDeliveryLogConsumer(ctx context.Context, db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}
		msgs, err := readDeliveryMessages(ctx, rdb)
		if err != nil {
			if err != platformredis.ErrNil {
				logger.Warn("delivery log consumer: read failed", "error", err)
				time.Sleep(time.Second)
			}
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		ids, err := processDeliveryLogBatch(ctx, db, msgs, logger)
		if err != nil {
			logger.Error("delivery log consumer: batch failed", "count", len(msgs), "error", err)
			continue
		}
		if err := rdb.XAck(ctx, billingEventsStream, deliveryLogProcessorGroup, ids...); err != nil {
			logger.Error("delivery log consumer: xack failed", "ids", len(ids), "error", err)
		}
	}
}

func readDeliveryMessages(ctx context.Context, rdb *platformredis.RedisClient) ([]platformredis.StreamMessage, error) {
	msgs, err := rdb.XReadGroup(
		ctx,
		deliveryLogProcessorGroup,
		deliveryLogConsumer,
		billingEventsStream,
		"0",
		deliveryReadBatchSize,
		0,
	)
	if err != nil && err != platformredis.ErrNil {
		return nil, err
	}
	if len(msgs) > 0 {
		return msgs, nil
	}
	return rdb.XReadGroup(
		ctx,
		deliveryLogProcessorGroup,
		deliveryLogConsumer,
		billingEventsStream,
		">",
		deliveryReadBatchSize,
		deliveryReadBlock,
	)
}

func processDeliveryLogBatch(
	ctx context.Context,
	db *gorm.DB,
	msgs []platformredis.StreamMessage,
	logger *slog.Logger,
) ([]string, error) {
	built := make([]deliverydomain.DeliveryLog, 0, len(msgs))
	ids := make([]string, 0, len(msgs))
	secret := strings.TrimSpace(os.Getenv("CLICK_TOKEN_SECRET"))

	for _, msg := range msgs {
		logRow, err := deliveryApp.MapStreamToDeliveryLog(deliveryApp.StreamDeliveryFields{
			CampaignID:     msg.Values["campaign_id"],
			EventType:      msg.Values["event_type"],
			PseudonymousID: msg.Values["pseudonymous_id"],
			Source:         msg.Values["source"],
			InstallToken:   msg.Values["install_token"],
			OccurredAt:     msg.Values["occurred_at"],
		}, secret)
		if err != nil {
			logger.Warn("delivery log consumer: skip message",
				"id", msg.ID,
				"campaign_id", msg.Values["campaign_id"],
				"error", err,
			)
			ids = append(ids, msg.ID)
			continue
		}
		built = append(built, *logRow)
		ids = append(ids, msg.ID)
	}

	if len(built) > 0 {
		logRepo := deliveryInfra.NewDeliveryLogRepository(db)
		if err := logRepo.CreateBatch(ctx, built); err != nil {
			return nil, err
		}
		jobRepo := deliveryInfra.NewDeliveryRepository(db)
		for i := range built {
			if err := jobRepo.RecordJob(ctx, built[i].PseudonymousID, built[i].CampaignID); err != nil {
				logger.Warn("delivery log consumer: delivery_jobs insert failed",
					"campaign_id", built[i].CampaignID,
					"pseudonymous_id", built[i].PseudonymousID,
					"error", err,
				)
			}
		}
	}
	return ids, nil
}
