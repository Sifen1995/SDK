package worker

import (
	"context"
	"log/slog"
	"strings"
	"time"

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
	logger.Info("delivery log consumer: group ready",
		"stream", billingEventsStream,
		"group", deliveryLogProcessorGroup,
	)
	go runDeliveryLogConsumer(context.Background(), db, rdb, logger)
}

func runDeliveryLogConsumer(ctx context.Context, db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}

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
			logger.Warn("delivery log consumer: xreadgroup pending failed", "error", err)
			time.Sleep(time.Second)
			continue
		}
		if len(msgs) == 0 {
			msgs, err = rdb.XReadGroup(
				ctx,
				deliveryLogProcessorGroup,
				deliveryLogConsumer,
				billingEventsStream,
				">",
				deliveryReadBatchSize,
				deliveryReadBlock,
			)
			if err != nil {
				if err == platformredis.ErrNil {
					continue
				}
				logger.Warn("delivery log consumer: xreadgroup failed", "error", err)
				time.Sleep(time.Second)
				continue
			}
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
			continue
		}
		logger.Info("delivery log consumer: processed batch", "count", len(ids))
	}
}

func processDeliveryLogBatch(
	ctx context.Context,
	db *gorm.DB,
	msgs []platformredis.StreamMessage,
	logger *slog.Logger,
) ([]string, error) {
	built := make([]deliverydomain.DeliveryLog, 0, len(msgs))
	ids := make([]string, 0, len(msgs))
	skipped := 0

	for _, msg := range msgs {
		logRow, err := mapStreamToDeliveryLog(msg)
		if err != nil {
			skipped++
			logger.Warn("delivery log consumer: skip message",
				"id", msg.ID,
				"campaign_id", strings.TrimSpace(msg.Values["campaign_id"]),
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
			if err := jobRepo.RecordJob(ctx, built[i].UserID, built[i].CampaignID); err != nil {
				logger.Warn("delivery log consumer: delivery_jobs insert failed",
					"campaign_id", built[i].CampaignID,
					"user_id", built[i].UserID,
					"error", err,
				)
			}
		}
	}
	if skipped > 0 {
		logger.Warn("delivery log consumer: skipped invalid messages",
			"skipped", skipped,
			"accepted", len(built),
		)
	}
	return ids, nil
}

func mapStreamToDeliveryLog(msg platformredis.StreamMessage) (*deliverydomain.DeliveryLog, error) {
	campaignID := strings.TrimSpace(msg.Values["campaign_id"])
	eventType := strings.ToLower(strings.TrimSpace(msg.Values["event_type"]))
	if campaignID == "" || eventType == "" {
		return nil, errRequiredFields
	}

	status := statusForEvent(eventType)
	if status == "" {
		return nil, errUnsupportedEvent
	}

	userID := strings.TrimSpace(msg.Values["pseudonymous_id"])
	sessionID := "telemetry"
	if userID == "" {
		userID = deliverydomain.AnonymousUserID
		sessionID = "anonymous"
	}
	if src := strings.TrimSpace(msg.Values["source"]); src == "anonymous" {
		userID = deliverydomain.AnonymousUserID
		sessionID = "anonymous"
	}

	return &deliverydomain.DeliveryLog{
		CampaignID:     campaignID,
		UserID:         userID,
		SessionID:      sessionID,
		DeliveryStatus: status,
		LoggedAt:       parseOccurredAt(msg.Values["occurred_at"]),
	}, nil
}

func statusForEvent(eventType string) string {
	switch eventType {
	case "impression":
		return deliverydomain.StatusRendered
	case "click":
		return deliverydomain.StatusClicked
	case "install", "signup", "purchase":
		return deliverydomain.StatusConverted
	default:
		return ""
	}
}

func parseOccurredAt(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC()
	}
	return time.Now().UTC()
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

const (
	errRequiredFields   = simpleError("campaign_id and event_type are required")
	errUnsupportedEvent = simpleError("unsupported event_type for delivery log")
)
