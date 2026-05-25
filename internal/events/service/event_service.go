package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"skykin-platform/internal/common/websocket"
	"skykin-platform/internal/events/dto"
	eventsModel "skykin-platform/internal/events/model"
	eventsRepo "skykin-platform/internal/events/repository"
	"skykin-platform/internal/intents/mlclient"
	intentModel "skykin-platform/internal/intents/model"
	intentsRepo "skykin-platform/internal/intents/repository"
	rewardModel "skykin-platform/internal/rewards/model"
	rewardsRepo "skykin-platform/internal/rewards/repository"
	usersRepo "skykin-platform/internal/users/repository"
)

type EventServiceInterface interface {
	ProcessEvent(ctx context.Context, extUserID string, event *eventsModel.Event) (*dto.EventResponseDTO, int, error)
}

type EventService struct {
	repo       eventsRepo.EventRepository
	userRepo   usersRepo.UserRepository // Remains injected cleanly here
	mlClient   *mlclient.MLClient
	rewardRepo rewardsRepo.RewardRepository
	intentRepo intentsRepo.IntentRepository
	notifier   websocket.Notifier
}

func NewEventService(
	r eventsRepo.EventRepository,
	ur usersRepo.UserRepository,
	mlc *mlclient.MLClient,
	rr rewardsRepo.RewardRepository,
	ir intentsRepo.IntentRepository,
	n websocket.Notifier,
) EventServiceInterface {
	return &EventService{
		repo:       r,
		userRepo:   ur,
		mlClient:   mlc,
		rewardRepo: rr,
		intentRepo: ir,
		notifier:   n,
	}
}

func (s *EventService) ProcessEvent(ctx context.Context, extUserID string, event *eventsModel.Event) (*dto.EventResponseDTO, int, error) {
	// 1. Let the User Module's Repository encapsulate the Find-or-Create loop
	user, err := s.userRepo.FindOrCreate(ctx, extUserID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to handle user matching context: %w", err)
	}

	// 2. Link the internal tracking UUID directly to the ingestion event payload
	event.UserID = user.ID

	// 3. Persist the event safely to the core tracking ledger
	if err := s.repo.Create(ctx, event); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save event: %w", err)
	}

	// 4. Fetch recent event history for context window processing
	history, err := s.repo.GetRecentEvents(ctx, event.UserID, 5)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to load event history: %w", err)
	}

	// 5. Cold start check: require more than two historical data footprints before running ML evaluations
	if len(history) <= 2 {
		return &dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			QueuedForAnalysis: true,
			RewardTriggered:   false,
		}, http.StatusAccepted, nil
	}

	// 6. Chronologically align events from oldest → newest for ML ingestion sequence compatibility
	chronological := make([]eventsModel.Event, len(history))
	for i := range history {
		chronological[len(history)-1-i] = history[i]
	}
	mlResult, err := s.mlClient.PredictIntent(extUserID, chronological)
	if err != nil {
		// Fallback gracefully if ML microservice is unavailable so ingestion processing isn't blocked
		return &dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			QueuedForAnalysis: true,
			RewardTriggered:   false,
		}, http.StatusAccepted, nil
	}

	// 7. Persist predicted Intent record to Postgres
	intentRecord := &intentModel.Intent{
		UserID:     user.ID,
		IntentName: mlResult.Intent,
		Confidence: mlResult.Confidence,
		CreatedAt:  time.Now(),
	}
	intent, err := s.intentRepo.Create(ctx, intentRecord)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save intent logs: %w", err)
	}

	// 8. Condition A: Intent identified, but confidence value misses the targeting reward criteria
	if !mlResult.RewardTriggered {
		response := dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			Intent:            intent.IntentName,
			Confidence:        intent.Confidence,
			RewardTriggered:   false,
			QueuedForAnalysis: false,
		}
		return &response, http.StatusCreated, nil
	}

	// 9. Condition B: Reward evaluation triggered -> Look up active matching profile rule
	rule, err := s.rewardRepo.GetRuleByIntent(ctx, mlResult.Intent)
	if err != nil {
		// Fallback gracefully: return intent info safely even if configuration rule is inactive
		return &dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			Intent:            intent.IntentName,
			Confidence:        intent.Confidence,
			RewardTriggered:   false,
			QueuedForAnalysis: false,
		}, http.StatusCreated, nil
	}

	// 10. Generate and assign the core Reward allocation payload metrics
	reward := &rewardModel.Reward{
		UserID:     user.ID,
		IntentID:   intent.ID,
		RuleID:     rule.ID,
		RewardType: rule.RewardType,
		Amount:     rule.Amount,
		Currency:   rule.Currency,
		Status:     "pending",
		Message:    rule.Message,
		CreatedAt:  time.Now(),
	}

	if err := s.rewardRepo.CreateReward(ctx, reward); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to persist reward allocation: %w", err)
	}

	// Asynchronously fire notification over WebSocket workers
	go s.notifier.NotifyUser(user.ID, reward)

	// 11. Return full confirmation payload envelope detailing the newly issued reward allocation
	return &dto.EventResponseDTO{
		EventID:         event.ID,
		UserID:          extUserID,
		Intent:          intent.IntentName,
		Confidence:      intent.Confidence,
		RewardTriggered: true,
		Reward: &dto.RewardResponseDTO{
			RewardID:   reward.ID,
			RewardType: reward.RewardType,
			Amount:     reward.Amount,
			Currency:   reward.Currency,
			Message:    reward.Message,
		},
		QueuedForAnalysis: false,
	}, http.StatusCreated, nil
}
