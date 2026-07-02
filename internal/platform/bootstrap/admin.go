package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"skykin-platform/configs"
	adminApp "skykin-platform/internal/admin/application"
	intentdomain "skykin-platform/internal/intents/domain"
	usersdomain "skykin-platform/internal/users/domain"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	usersInfra "skykin-platform/internal/users/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type intentBatchFetcherAdapter struct {
	repo intentdomain.IntentRepository
}

func (a *intentBatchFetcherAdapter) FindLatestByUserIDs(
	ctx context.Context,
	userIDs []uuid.UUID,
) (map[uuid.UUID]*adminApp.IntentSummary, error) {
	raw, err := a.repo.FindLatestByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*adminApp.IntentSummary, len(raw))
	for userID, intent := range raw {
		result[userID] = &adminApp.IntentSummary{
			IntentName:  intent.IntentName,
			Confidence:  intent.Confidence,
			PredictedAt: intent.CreatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

type userListerAdapter struct {
	repo usersdomain.UserRepository
}

func (a *userListerAdapter) FindAll(
	ctx context.Context,
	limit int,
	offset int,
) ([]*adminApp.UserSummary, int64, error) {
	users, total, err := a.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*adminApp.UserSummary, len(users))
	for i, u := range users {
		id, err := uuid.Parse(u.ID)
		if err != nil {
			return nil, 0, err
		}
		result[i] = &adminApp.UserSummary{
			ID:             id,
			ExternalUserID: u.ExternalUserID,
			CreatedAt:      u.CreatedAt,
		}
	}
	return result, total, nil
}

// NewGetUsersWithIntentsUseCase wires users/intents infrastructure into the admin use case
// via adapters that map foreign domain types to admin-owned types.
func NewGetUsersWithIntentsUseCase(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *adminApp.GetUsersWithIntentsUseCase {
	userRepo := usersInfra.NewUserRepository(db)
	intentRepo := intentsInfra.NewIntentRepository(db, cfg)

	return adminApp.NewGetUsersWithIntentsUseCase(
		&userListerAdapter{repo: userRepo},
		&intentBatchFetcherAdapter{repo: intentRepo},
		logger,
	)
}
