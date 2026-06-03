package consumers

import (
	"context"
	"log/slog"
	"strings"
	"time"

	internalevents "skykin-platform/internal/events/events"
	"skykin-platform/internal/intents/application"
	"skykin-platform/internal/platform/messaging"
	platformredis "skykin-platform/internal/platform/redis"

	"github.com/redis/go-redis/v9"
)

const userPipelineKey = "user_pipeline"

// PredictIntentConsumer triggers intent prediction asynchronously.
type PredictIntentConsumer struct {
	predict *application.PredictIntentUseCase
	redis   *platformredis.RedisClient
	logger  *slog.Logger
}

func NewPredictIntentConsumer(
	predict *application.PredictIntentUseCase,
	redisClient *platformredis.RedisClient,
) *PredictIntentConsumer {
	return &PredictIntentConsumer{
		predict: predict,
		redis:   redisClient,
		logger:  slog.Default(),
	}
}

// Register starts async prediction triggers.
func (c *PredictIntentConsumer) Register(bus *messaging.Bus) {
	// Primary path: in-process bus (reliable in monolith).
	messaging.Register(bus, internalevents.TopicIntentEvaluationRequested, c.handleBusEvent)

	// Secondary path: Redis queue worker (for future scaling).
	if c.redis != nil {
		go c.runWorker(context.Background())
	}
}

func (c *PredictIntentConsumer) handleBusEvent(e messaging.Event) {
	payload, ok := e.Payload.(internalevents.IntentEvaluationRequested)
	if !ok {
		return
	}
	c.runPrediction(e.Ctx, payload.UserID)
}

func (c *PredictIntentConsumer) runWorker(ctx context.Context) {
	c.logger.Info("intent prediction redis worker started")
	for {
		userID, err := c.redis.BRPop(ctx, userPipelineKey, 5*time.Second)
		if err != nil {
			if err == redis.Nil {
				continue
			}
			c.logger.Warn("redis worker BRPOP failed", "error", err)
			time.Sleep(time.Second)
			continue
		}
		if strings.TrimSpace(userID) == "" {
			continue
		}
		c.runPrediction(ctx, userID)
	}
}

func (c *PredictIntentConsumer) runPrediction(ctx context.Context, externalUserID string) {
	if ctx == nil {
		ctx = context.Background()
	}
	result, err := c.predict.Execute(ctx, externalUserID)
	if err != nil {
		c.logger.Warn("prediction failed", "user_id", externalUserID, "error", err)
		return
	}
	if result == nil {
		return
	}
	if result.Status != "predicted" {
		c.logger.Info("prediction skipped",
			"user_id", externalUserID,
			"status", result.Status,
			"message", result.Message,
		)
	}
}
