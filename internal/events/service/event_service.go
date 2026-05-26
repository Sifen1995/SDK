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
	ProcessEvent(ctx context.Context, appID string, req *dto.EventRequestDTO) (*dto.EventResponseDTO, int, error)
	ProcessBatchEvents(ctx context.Context, appID string, req *dto.BatchEventRequestDTO) (*dto.BatchEventResponseDTO, int, error)
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

func (s *EventService) ProcessEvent(ctx context.Context, appID string, req *dto.EventRequestDTO) (*dto.EventResponseDTO, int, error) {
	// 1. Resolve or establish the user track record using external_user_id inside the DTO object
	user, err := s.userRepo.FindOrCreate(ctx, req.ExternalUserID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to handle user matching context: %w", err)
	}

	// 2. Map metrics into the database GORM Model instance layout
	// Note: If you want to segment events by app later, you can add an ApplicationID column to your events model/table
	event := &eventsModel.Event{
		EventID:    req.EventID,     
		EventType:  req.EventType,   
		Metadata:   req.Metadata,    
		Timestamp:  req.Timestamp,   
		UserID:     user.ID,         
		IsFlagged:  false,           
		ReceivedAt: time.Now(),      
	}

	// 3. Persist the event safely to the database
	if err := s.repo.Create(ctx, event); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save event: %w", err)
	}

	// 4. Fetch recent event history for context window processing
	history, err := s.repo.GetRecentEvents(ctx, event.UserID, 5)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to load event history: %w", err)
	}

	// 5. Cold start check
	if len(history) <= 2 {
		return &dto.EventResponseDTO{
			EventID:           event.EventID,
			UserID:            req.ExternalUserID, // Now safely reflects client tracking id
			QueuedForAnalysis: true,
			RewardTriggered:   false,
		}, http.StatusAccepted, nil
	}

	// 6. Chronologically align events for ML ingestion
	chronological := make([]eventsModel.Event, len(history))
	for i := range history {
		chronological[len(history)-1-i] = history[i]
	}
	
	// Call ML model processing passing client user profile identity context
	mlResult, err := s.mlClient.PredictIntent(req.ExternalUserID, chronological)
	if err != nil {
		return &dto.EventResponseDTO{
			EventID:           event.EventID,
			UserID:            req.ExternalUserID,
			QueuedForAnalysis: true,
			RewardTriggered:   false,
		}, http.StatusAccepted, nil
	}

	// 7. Persist predicted Intent record
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

	// 8. Condition A: Intent identified, but confidence value misses the threshold
	if !mlResult.RewardTriggered {
		response := dto.EventResponseDTO{
			EventID:           event.EventID,
			UserID:            req.ExternalUserID,
			Intent:            intent.IntentName,
			Confidence:        intent.Confidence,
			RewardTriggered:   false,
			QueuedForAnalysis: false,
		}
		return &response, http.StatusCreated, nil
	}

	// 9. Condition B: Reward evaluation triggered
	rule, err := s.rewardRepo.GetRuleByIntent(ctx, mlResult.Intent)
	if err != nil {
		return &dto.EventResponseDTO{
			EventID:           event.EventID,
			UserID:            req.ExternalUserID,
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

	go s.notifier.NotifyUser(req.ExternalUserID, map[string]interface{}{
		"type":        "reward_earned",
		"reward_id":   reward.ID,
		"reward_type": reward.RewardType,
		"amount":      reward.Amount,
		"currency":    reward.Currency,
		"message":     reward.Message,
		"intent":      intent.IntentName,
		"confidence":  intent.Confidence,
		"created_at":  reward.CreatedAt,
	})

	// 11. Return full confirmation payload envelope
	return &dto.EventResponseDTO{
		EventID:         event.EventID,
		UserID:          req.ExternalUserID,
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

func (s *EventService) ProcessBatchEvents(ctx context.Context, appID string, req *dto.BatchEventRequestDTO) (*dto.BatchEventResponseDTO, int, error) {
	user, err := s.userRepo.FindOrCreate(ctx, req.ExternalUserID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to resolve user: %w", err)
	}

	now := time.Now()
	models := make([]eventsModel.Event, len(req.Events))
	for i, e := range req.Events {
		meta := e.Metadata
		if meta == nil {
			meta = map[string]interface{}{}
		}
		models[i] = eventsModel.Event{
			EventID:    e.EventID,
			EventType:  e.EventType,
			Metadata:   meta,
			Timestamp:  e.Timestamp,
			UserID:     user.ID,
			IsFlagged:  false,
			ReceivedAt: now,
		}
	}

	if err := s.repo.CreateBatch(ctx, models); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save events batch: %w", err)
	}

	resp := &dto.BatchEventResponseDTO{
		UserID:         req.ExternalUserID,
		EventsReceived: len(req.Events),
	}

	history, err := s.repo.GetRecentEvents(ctx, user.ID, 5)
	if err != nil || len(history) <= 2 {
		return resp, http.StatusAccepted, nil
	}

	chronological := make([]eventsModel.Event, len(history))
	for i := range history {
		chronological[len(history)-1-i] = history[i]
	}

	mlResult, err := s.mlClient.PredictIntent(req.ExternalUserID, chronological)
	if err != nil {
		return resp, http.StatusAccepted, nil
	}

	intentRecord := &intentModel.Intent{
		UserID:     user.ID,
		IntentName: mlResult.Intent,
		Confidence: mlResult.Confidence,
		CreatedAt:  time.Now(),
	}
	intent, err := s.intentRepo.Create(ctx, intentRecord)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save intent: %w", err)
	}

	resp.Intent = intent.IntentName
	resp.Confidence = intent.Confidence

	if !mlResult.RewardTriggered {
		return resp, http.StatusCreated, nil
	}

	rule, err := s.rewardRepo.GetRuleByIntent(ctx, mlResult.Intent)
	if err != nil {
		return resp, http.StatusCreated, nil
	}

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
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to persist reward: %w", err)
	}

	resp.RewardTriggered = true
	resp.Reward = &dto.RewardResponseDTO{
		RewardID:   reward.ID,
		RewardType: reward.RewardType,
		Amount:     reward.Amount,
		Currency:   reward.Currency,
		Message:    reward.Message,
	}

	go s.notifier.NotifyUser(req.ExternalUserID, map[string]interface{}{
		"type":        "reward_earned",
		"reward_id":   reward.ID,
		"reward_type": reward.RewardType,
		"amount":      reward.Amount,
		"currency":    reward.Currency,
		"message":     reward.Message,
		"intent":      intent.IntentName,
		"confidence":  intent.Confidence,
		"created_at":  reward.CreatedAt,
	})

	return resp, http.StatusCreated, nil
}