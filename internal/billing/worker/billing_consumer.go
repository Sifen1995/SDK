package worker

import (
	"context"
	"log/slog"
	"strings"
	"time"

	billingApp "skykin-platform/internal/billing/application"
	platformredis "skykin-platform/internal/platform/redis"
)

const (
	billingEventsStream      = "stream:billing_events"
	billingProcessorGroup    = "billing_processor_group"
	billingProcessorConsumer = "billing-worker-1"
	billingReadBatchSize     = 100
	billingReadBlock         = 2 * time.Second
)

// StartBillingConsumer owns Redis transport only. Billing decisions and writes
// are delegated to the application processor supplied by the composition root.
func StartBillingConsumer(
	processor *billingApp.EventProcessor,
	rdb *platformredis.RedisClient,
	logger *slog.Logger,
) {
	if processor == nil || rdb == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	if err := rdb.XGroupCreateMkStream(
		context.Background(),
		billingEventsStream,
		billingProcessorGroup,
		"0",
	); err != nil {
		logger.Error("billing consumer: create group failed", "error", err)
		return
	}
	go runBillingConsumer(context.Background(), processor, rdb, logger)
}

func runBillingConsumer(
	ctx context.Context,
	processor *billingApp.EventProcessor,
	rdb *platformredis.RedisClient,
	logger *slog.Logger,
) {
	for {
		if ctx.Err() != nil {
			return
		}
		msgs, err := readBillingMessages(ctx, rdb)
		if err != nil {
			if err != platformredis.ErrNil {
				logger.Warn("billing consumer: read failed", "error", err)
				time.Sleep(time.Second)
			}
			continue
		}
		if len(msgs) == 0 {
			continue
		}

		ids := processBillingBatch(ctx, processor, msgs, logger)
		if err := rdb.XAck(ctx, billingEventsStream, billingProcessorGroup, ids...); err != nil {
			logger.Error("billing consumer: xack failed", "ids", len(ids), "error", err)
		}
	}
}

func readBillingMessages(
	ctx context.Context,
	rdb *platformredis.RedisClient,
) ([]platformredis.StreamMessage, error) {
	msgs, err := rdb.XReadGroup(
		ctx,
		billingProcessorGroup,
		billingProcessorConsumer,
		billingEventsStream,
		"0",
		billingReadBatchSize,
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
		billingProcessorGroup,
		billingProcessorConsumer,
		billingEventsStream,
		">",
		billingReadBatchSize,
		billingReadBlock,
	)
}

func processBillingBatch(
	ctx context.Context,
	processor *billingApp.EventProcessor,
	msgs []platformredis.StreamMessage,
	logger *slog.Logger,
) []string {
	ids := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		err := processor.Process(ctx, billingApp.BillingInput{
			CampaignID:     strings.TrimSpace(msg.Values["campaign_id"]),
			EventType:      strings.TrimSpace(msg.Values["event_type"]),
			TransactionRaw: strings.TrimSpace(msg.Values["transaction_value"]),
			OccurredAtRaw:  strings.TrimSpace(msg.Values["occurred_at"]),
		})
		if err != nil {
			logger.Warn(
				"billing consumer: skip message",
				"id", msg.ID,
				"campaign_id", msg.Values["campaign_id"],
				"event_type", msg.Values["event_type"],
				"error", err,
			)
		}
		// Invalid messages are acknowledged as poison messages. Persistence
		// failures remain visible in logs but are also acknowledged to prevent
		// an unbounded pending loop; a durable retry/DLQ can be added separately.
		ids = append(ids, msg.ID)
	}
	return ids
}
