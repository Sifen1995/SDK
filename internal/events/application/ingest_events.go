package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"skykin-platform/internal/events/domain"
	internalevents "skykin-platform/internal/events/events"
	platformredis "skykin-platform/internal/platform/redis"
)

const statusStored = "stored"
const statusDuplicate = "duplicate"
const statusInvalid = "invalid"

// EventInput is the application-layer representation of one SDK event.
type EventInput struct {
	EventID    string
	EventType  string
	Domain     string
	SessionID  string
	ScreenName string
	Metadata   map[string]any
	DeviceType string
	Platform   string
	AppVersion string
	CreatedAt  time.Time
}

// IngestCommand carries a batch ingestion request.
type IngestCommand struct {
	ApplicationID  string
	ExternalUserID string
	Events         []EventInput
}

// EventIngestResult describes per-event processing outcome.
type EventIngestResult struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

// IngestResult is returned to the HTTP layer.
type IngestResult struct {
	Accepted          bool                `json:"accepted"`
	PredictionQueued  bool                `json:"prediction_queued"`
	Results           []EventIngestResult `json:"results"`
}

// IngestEventsUseCase orchestrates validation, deduplication, persistence, and publishing.
type IngestEventsUseCase struct {
	repo      domain.EventRepository
	users     UserResolver
	dedup     DedupStore
	publisher EventPublisher
	logger    *slog.Logger
	redis     *platformredis.RedisClient
}

func NewIngestEventsUseCase(
	repo domain.EventRepository,
	users UserResolver,
	dedup DedupStore,
	redisClient *platformredis.RedisClient,
	publisher EventPublisher,
	logger *slog.Logger,
) *IngestEventsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &IngestEventsUseCase{
		repo:      repo,
		users:     users,
		dedup:     dedup,
		publisher: publisher,
		logger:    logger,
		redis:     redisClient,
	}
}

func (uc *IngestEventsUseCase) Execute(ctx context.Context, cmd IngestCommand) (*IngestResult, error) {
	if cmd.ExternalUserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if len(cmd.Events) == 0 {
		return nil, fmt.Errorf("events batch cannot be empty")
	}

	user, err := uc.users.FindOrCreate(ctx, cmd.ExternalUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	results := make([]EventIngestResult, 0, len(cmd.Events))
	toStore := make([]domain.Event, 0, len(cmd.Events))
	now := time.Now().UTC()

	for _, input := range cmd.Events {
		if err := ValidateEventInput(
			input.EventID,
			input.EventType,
			input.Domain,
			input.SessionID,
			input.ScreenName,
			input.Metadata,
		); err != nil {
			uc.logger.Warn("event validation failed",
				"event_id", input.EventID,
				"event_type", input.EventType,
				"user_id", cmd.ExternalUserID,
				"status", statusInvalid,
				"error", err,
			)
			results = append(results, EventIngestResult{EventID: input.EventID, Status: statusInvalid})
			continue
		}

		dup, err := IsDuplicate(ctx, uc.dedup, input.EventID)
		if err != nil {
			return nil, fmt.Errorf("dedup check failed for event %s: %w", input.EventID, err)
		}
		if dup {
			uc.logger.Info("duplicate event skipped",
				"event_id", input.EventID,
				"event_type", input.EventType,
				"user_id", cmd.ExternalUserID,
				"status", statusDuplicate,
			)
			results = append(results, EventIngestResult{EventID: input.EventID, Status: statusDuplicate})
			continue
		}

		createdAt := input.CreatedAt
		if createdAt.IsZero() {
			createdAt = now
		}

		event := domain.Event{
			EventID:       input.EventID,
			UserID:        user.ID,
			ApplicationID: cmd.ApplicationID,
			EventType:     domain.EventType(input.EventType),
			Domain:        input.Domain,
			SessionID:     input.SessionID,
			ScreenName:    input.ScreenName,
			Metadata:      input.Metadata,
			DeviceType:    input.DeviceType,
			Platform:      input.Platform,
			AppVersion:    input.AppVersion,
			CreatedAt:     createdAt,
		}
		toStore = append(toStore, event)
		results = append(results, EventIngestResult{EventID: input.EventID, Status: statusStored})
	}

	if len(toStore) == 0 {
		return &IngestResult{Accepted: true, PredictionQueued: false, Results: results}, nil
	}

	if err := uc.repo.SaveBatch(ctx, toStore); err != nil {
		return nil, fmt.Errorf("save batch: %w", err)
	}

	for _, ev := range toStore {
		uc.logger.Info("event stored",
			"event_id", ev.EventID,
			"event_type", string(ev.EventType),
			"user_id", cmd.ExternalUserID,
			"status", statusStored,
		)

		uc.publisher.Publish(ctx, internalevents.TopicEventStored, internalevents.EventStored{
			EventID:       ev.EventID,
			UserID:        cmd.ExternalUserID,
			ApplicationID: ev.ApplicationID,
			EventType:     string(ev.EventType),
			Domain:        ev.Domain,
			SessionID:     ev.SessionID,
		})

		uc.publisher.Publish(ctx, internalevents.TopicEventReceived, internalevents.EventReceived{
			EventID:       ev.EventID,
			UserID:        cmd.ExternalUserID,
			ApplicationID: ev.ApplicationID,
			EventType:     string(ev.EventType),
			Domain:        ev.Domain,
			SessionID:     ev.SessionID,
		})
	}

	if uc.redis != nil {
		if err := uc.cacheUserEvents(ctx, cmd.ExternalUserID, toStore); err != nil {
			uc.logger.Warn("failed to cache user events in redis",
				"user_id", cmd.ExternalUserID,
				"error", err,
			)
		}
	}

	return &IngestResult{
		Accepted:         true,
		PredictionQueued: false,
		Results:          results,
	}, nil
}

func (uc *IngestEventsUseCase) cacheUserEvents(ctx context.Context, externalUserID string, events []domain.Event) error {
	key := "user_events:" + externalUserID
	values := make([]string, 0, len(events))
	for _, ev := range events {
		raw, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", ev.EventID, err)
		}
		values = append(values, string(raw))
	}
	if len(values) == 0 {
		return nil
	}
	if err := uc.redis.LPush(ctx, key, values...); err != nil {
		return err
	}
	// keep last 200 events per user for inference context
	if err := uc.redis.LTrim(ctx, key, 0, 199); err != nil {
		return err
	}
	// expire inactive users after 24h
	if err := uc.redis.Expire(ctx, key, 24*time.Hour); err != nil {
		return err
	}
	return nil
}
