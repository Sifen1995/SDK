package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type UserWithIntent struct {
	UserID         uuid.UUID      `json:"user_id"`
	ExternalUserID string         `json:"external_user_id"`
	CreatedAt      string         `json:"created_at"`
	LatestIntent   *IntentSummary `json:"latest_intent,omitempty"`
}

type IntentSummary struct {
	IntentName  string  `json:"intent_name"`
	Confidence  float64 `json:"confidence"`
	PredictedAt string  `json:"predicted_at"`
}

// IntentBatchFetcher returns admin-owned intent summaries keyed by user id.
type IntentBatchFetcher interface {
	FindLatestByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*IntentSummary, error)
}

// UserLister returns paginated admin-owned user summaries.
type UserLister interface {
	FindAll(ctx context.Context, limit, offset int) ([]*UserSummary, int64, error)
}

type UserSummary struct {
	ID             uuid.UUID
	ExternalUserID string
	CreatedAt      time.Time
}

type GetUsersRequest struct {
	Page    int
	PerPage int
}

type GetUsersResult struct {
	Users      []*UserWithIntent `json:"users"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

type GetUsersWithIntentsUseCase struct {
	userRepo   UserLister
	intentRepo IntentBatchFetcher
	logger     *slog.Logger
}

func NewGetUsersWithIntentsUseCase(
	userRepo UserLister,
	intentRepo IntentBatchFetcher,
	logger *slog.Logger,
) *GetUsersWithIntentsUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetUsersWithIntentsUseCase{
		userRepo:   userRepo,
		intentRepo: intentRepo,
		logger:     logger,
	}
}

func (uc *GetUsersWithIntentsUseCase) Execute(
	ctx context.Context,
	req GetUsersRequest,
) (*GetUsersResult, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PerPage < 1 {
		req.PerPage = 20
	}
	if req.PerPage > 100 {
		req.PerPage = 100
	}

	offset := (req.Page - 1) * req.PerPage
	users, total, err := uc.userRepo.FindAll(ctx, req.PerPage, offset)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, len(users))
	for i, u := range users {
		userIDs[i] = u.ID
	}

	intentMap, err := uc.intentRepo.FindLatestByUserIDs(ctx, userIDs)
	if err != nil {
		uc.logger.Warn("intent fetch failed, returning users only", "error", err)
		intentMap = map[uuid.UUID]*IntentSummary{}
	}

	result := make([]*UserWithIntent, len(users))
	for i, u := range users {
		uwi := &UserWithIntent{
			UserID:         u.ID,
			ExternalUserID: u.ExternalUserID,
			CreatedAt:      u.CreatedAt.Format(time.RFC3339),
		}
		if intent, ok := intentMap[u.ID]; ok {
			uwi.LatestIntent = intent
		}
		result[i] = uwi
	}

	totalPages := 0
	if req.PerPage > 0 && total > 0 {
		totalPages = int(total) / req.PerPage
		if int(total)%req.PerPage != 0 {
			totalPages++
		}
	}

	return &GetUsersResult{
		Users:      result,
		Total:      total,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalPages: totalPages,
	}, nil
}
