package bootstrap

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"skykin-platform/configs"
	adminApp "skykin-platform/internal/admin/application"
	intentdomain "skykin-platform/internal/intents/domain"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	usersdomain "skykin-platform/internal/users/domain"
	usersInfra "skykin-platform/internal/users/infrastructure"

	"gorm.io/gorm"
)

type intentBatchFetcherAdapter struct {
	repo intentdomain.IntentRepository
}

func (a *intentBatchFetcherAdapter) FindLatestByUserIDs(
	ctx context.Context,
	userIDs []int64,
) (map[int64]*adminApp.IntentSummary, error) {
	ids := make([]string, len(userIDs))
	for i, id := range userIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}

	raw, err := a.repo.FindLatestByUserIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*adminApp.IntentSummary, len(raw))
	for sid, intent := range raw {
		uid, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			continue
		}
		result[uid] = &adminApp.IntentSummary{
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
		result[i] = &adminApp.UserSummary{
			ID:        u.ID,
			CreatedAt: u.CreatedAt,
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
