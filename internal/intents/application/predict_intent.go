package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"skykin-platform/internal/events/domain"
	eventsApp "skykin-platform/internal/events/application"
	intentdomain "skykin-platform/internal/intents/domain"
	intentEvents "skykin-platform/internal/intents/events"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	"skykin-platform/internal/platform/messaging"
	platformredis "skykin-platform/internal/platform/redis"
	campaignApp "skykin-platform/internal/campaigns/application"
	campaignEvents "skykin-platform/internal/campaigns/events"
)

const eventHistoryLimit = 200
const predictLockTTL = 30 * time.Second

// PredictIntentResult is the outcome of intent + optional reward evaluation.
type PredictIntentResult struct {
	UserID          string   `json:"user_id"`
	Status          string   `json:"status"`
	Intent          string   `json:"intent,omitempty"`
	Confidence      float64  `json:"confidence,omitempty"`
	RewardTriggered bool     `json:"reward_triggered"`
	TopSignals      []string `json:"top_signals,omitempty"`
	Reward          *Reward  `json:"reward,omitempty"`
	Message         string   `json:"message,omitempty"`
}

type Reward struct {
	RewardID   string  `json:"reward_id"`
	RewardType string  `json:"reward_type"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Message    string  `json:"message"`
}

// PredictIntentUseCase loads behavioral history and runs ML intent prediction.
type PredictIntentUseCase struct {
	eventRepo  domain.EventRepository
	users      eventsApp.UserResolver
	mlClient   *intentsInfra.MLClient
	intentRepo intentdomain.IntentRepository
	adDelivery *campaignApp.AdDeliveryService
	redis      *platformredis.RedisClient
	bus        *messaging.Bus
	logger     *slog.Logger
}

func NewPredictIntentUseCase(
	eventRepo domain.EventRepository,
	users eventsApp.UserResolver,
	mlClient *intentsInfra.MLClient,
	intentRepo intentdomain.IntentRepository,
	adDelivery *campaignApp.AdDeliveryService,
	redisClient *platformredis.RedisClient,
	bus *messaging.Bus,
	logger *slog.Logger,
) *PredictIntentUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &PredictIntentUseCase{
		eventRepo:  eventRepo,
		users:      users,
		mlClient:   mlClient,
		intentRepo: intentRepo,
		adDelivery: adDelivery,
		redis:      redisClient,
		bus:        bus,
		logger:     logger,
	}
}

// Execute runs prediction for one SDK user id.
func (uc *PredictIntentUseCase) Execute(ctx context.Context, externalUserID string) (*PredictIntentResult, error) {
	externalUserID = strings.TrimSpace(externalUserID)
	if externalUserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if uc.redis != nil {
		acquired, err := uc.redis.SetNX(ctx, predictLockKey(externalUserID), "1", predictLockTTL)
		if err != nil {
			return nil, fmt.Errorf("prediction lock: %w", err)
		}
		if !acquired {
			return &PredictIntentResult{
				UserID:  externalUserID,
				Status:  "skipped",
				Message: "prediction already in progress for this user",
			}, nil
		}
		defer func() {
			_ = uc.redis.Del(ctx, predictLockKey(externalUserID))
		}()
	}

	history, err := uc.loadHistory(ctx, externalUserID)
	if err != nil {
		return nil, err
	}
	if len(history) < 3 {
		return &PredictIntentResult{
			UserID:  externalUserID,
			Status:  "insufficient_history",
			Message: fmt.Sprintf("need at least 3 events, found %d", len(history)),
		}, nil
	}

	mlResult, err := uc.mlClient.PredictIntent(externalUserID, reverse(history))
	if err != nil {
		uc.logger.Warn("ml prediction failed", "user_id", externalUserID, "error", err)
		return &PredictIntentResult{
			UserID:  externalUserID,
			Status:  "ml_unavailable",
			Message: err.Error(),
		}, nil
	}

	if mlResult.Intent == "" {
		return &PredictIntentResult{
			UserID:  externalUserID,
			Status:  "insufficient_history",
			Message: "ml returned no intent for current session",
		}, nil
	}

	user, err := uc.users.FindOrCreate(ctx, externalUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve user: %w", err)
	}

	intent, err := uc.intentRepo.Create(ctx, &intentdomain.Intent{
		UserID:     user.ID,
		IntentName: mlResult.Intent,
		Confidence: mlResult.Confidence,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("save intent: %w", err)
	}

	result := &PredictIntentResult{
		UserID:          externalUserID,
		Status:          "predicted",
		Intent:          intent.IntentName,
		Confidence:      intent.Confidence,
		RewardTriggered: mlResult.RewardTriggered,
		TopSignals:      mlResult.TopSignals,
	}

	uc.notifyIntentPredicted(ctx, externalUserID, result)
	uc.deliverCampaignAd(ctx, externalUserID, user.ID, intent.IntentName, intent.Confidence, history)

	if mlResult.RewardTriggered && uc.bus != nil {
		uc.bus.Publish(messaging.Event{
			Name: intentEvents.TopicIntentRewardEligible,
			Ctx:  ctx,
			Payload: intentEvents.IntentRewardEligible{
				ExternalUserID: externalUserID,
				InternalUserID: user.ID,
				IntentID:       intent.ID,
				IntentName:     intent.IntentName,
			},
		})
	}

	uc.logger.Info("intent predicted",
		"user_id", externalUserID,
		"intent", result.Intent,
		"confidence", result.Confidence,
		"reward_triggered", result.RewardTriggered,
	)

	return result, nil
}

func (uc *PredictIntentUseCase) loadHistory(ctx context.Context, externalUserID string) ([]domain.Event, error) {
	if uc.redis != nil {
		key := userEventsKey(externalUserID)
		items, err := uc.redis.LRange(ctx, key, 0, eventHistoryLimit-1)
		if err == nil && len(items) > 0 {
			events := make([]domain.Event, 0, len(items))
			for _, raw := range items {
				var ev domain.Event
				if unmarshalErr := json.Unmarshal([]byte(raw), &ev); unmarshalErr != nil {
					continue
				}
				events = append(events, ev)
			}
			if len(events) > 0 {
				return events, nil
			}
		}
	}
	return uc.eventRepo.FindByUser(ctx, externalUserID, eventHistoryLimit)
}

func predictLockKey(externalUserID string) string {
	return "predict_lock:" + externalUserID
}

func userEventsKey(externalUserID string) string {
	return "user_events:" + externalUserID
}

func (uc *PredictIntentUseCase) notifyIntentPredicted(ctx context.Context, externalUserID string, result *PredictIntentResult) {
	if uc.bus == nil || result.Intent == "" {
		return
	}
	uc.bus.Publish(messaging.Event{
		Name: intentEvents.TopicIntentPredicted,
		Ctx:  ctx,
		Payload: intentEvents.IntentPredicted{
			ExternalUserID:  externalUserID,
			Intent:          result.Intent,
			Confidence:      result.Confidence,
			TopSignals:      result.TopSignals,
			RewardTriggered: result.RewardTriggered,
		},
	})
}

func (uc *PredictIntentUseCase) deliverCampaignAd(ctx context.Context, externalUserID, internalUserID, intentName string, confidence float64, history []domain.Event) {
	if uc.adDelivery == nil || uc.bus == nil {
		return
	}
	sessionID := ""
	if len(history) > 0 {
		sessionID = history[0].SessionID
	}
	ad, err := uc.adDelivery.BuildAdForIntent(ctx, intentName)
	if err != nil {
		uc.logger.Info("no campaign ad for intent", "user_id", externalUserID, "intent", intentName)
		return
	}
	payload := map[string]any{
		"type":            ad.Type,
		"intent":          ad.Intent,
		"confidence":      confidence,
		"campaign_id":     ad.CampaignID,
		"campaign_name":   ad.CampaignName,
		"channel_code":    ad.ChannelCode,
		"creative_format": ad.ChannelCode,
		"content":         ad.Content,
	}
	uc.bus.Publish(messaging.Event{
		Name: campaignEvents.TopicCampaignAdDelivered,
		Ctx:  ctx,
		Payload: campaignEvents.CampaignAdDelivered{
			ExternalUserID: externalUserID,
			InternalUserID: internalUserID,
			SessionID:      sessionID,
			Intent:         intentName,
			Ad:             payload,
		},
	})
	uc.logger.Info("campaign ad queued", "user_id", externalUserID, "intent", intentName, "channel", ad.ChannelCode)
}

func reverse(input []domain.Event) []domain.Event {
	out := make([]domain.Event, len(input))
	for i := range input {
		out[len(input)-1-i] = input[i]
	}
	return out
}
