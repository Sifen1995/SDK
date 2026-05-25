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
	intentModel "skykin-platform/internal/intents/model"
	intentsRepo "skykin-platform/internal/intents/repository"
	"skykin-platform/internal/intents/mlclient"
	rewardModel "skykin-platform/internal/rewards/model"
	rewardsRepo "skykin-platform/internal/rewards/repository"
	usersRepo "skykin-platform/internal/users/repository"
)

type EventServiceInterface interface {
	ProcessEvent(ctx context.Context, extUserID string, event *eventsModel.Event) (*dto.EventResponseDTO, int, error)
}

type EventService struct {
	repo       eventsRepo.EventRepository
	userRepo   usersRepo.UserRepository
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
	// 1. Find or Create User by external_user_id
	user, err := s.userRepo.FindOrCreate(ctx, extUserID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to handle user: %w", err)
	}

	// 2. Link the internal UUID to the event
	event.UserID = user.ID

	// 3. Persist the event safely
	if err := s.repo.Create(ctx, event); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save event: %w", err)
	}

	// 4. Fetch recent event history for context window
	history, err := s.repo.GetRecentEvents(ctx, event.UserID, 5)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to load event history: %w", err)
	}

	// 5. Cold start: need more than two historical events before calling ML
	if len(history) <= 2 {
		return &dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			QueuedForAnalysis: true,
			RewardTriggered:   false,
		}, http.StatusAccepted, nil
	}

	// 6. Call ML with events oldest → newest (repo returns newest first)
	chronological := make([]eventsModel.Event, len(history))
	for i := range history {
		chronological[len(history)-1-i] = history[i]
	}
	mlResult, err := s.mlClient.PredictIntent(extUserID, chronological)
	if err != nil {
		// Fallback gracefully if ML microservice is down so the SDK transaction doesn't break
		return &dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			QueuedForAnalysis: true,
			RewardTriggered:   false,
		}, http.StatusAccepted, nil
	}

	// 7. ALWAYS save the Intent record to Postgres
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

	// 8. Scenario A: Intent identified, but confidence DID NOT hit the reward threshold
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

	// 9. Scenario B: Reward IS triggered -> Look up cached rule from your internal repo
	rule, err := s.rewardRepo.GetRuleByIntent(ctx, mlResult.Intent)
	if err != nil {
		// If a rule is missing or inactive, still return the predicted intent details safely
		return &dto.EventResponseDTO{
			EventID:           event.ID,
			UserID:            extUserID,
			Intent:            intent.IntentName,
			Confidence:        intent.Confidence,
			RewardTriggered:   false,
			QueuedForAnalysis: false,
		}, http.StatusCreated, nil
	}

	// 10. Generate the Reward record mapping Intent ID and Rule ID
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
	go s.notifier.NotifyUser(user.ID, reward)

	// 11. Return full payload including the freshly minted reward data structure
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
