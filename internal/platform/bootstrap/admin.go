package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"skykin-platform/configs"
	adminApp "skykin-platform/internal/admin/application"
	consentdomain "skykin-platform/internal/consent/domain"
	consentInfra "skykin-platform/internal/consent/infrastructure"
	intentdomain "skykin-platform/internal/intents/domain"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	usersdomain "skykin-platform/internal/users/domain"
	usersInfra "skykin-platform/internal/users/infrastructure"

	"gorm.io/gorm"
)

// intentBatchFetcherAdapter bridges internal user ids to intent rows. Intents are
// keyed by pseudonymous id, so the mapping table is the only way to join the two.
type intentBatchFetcherAdapter struct {
	repo     intentdomain.IntentRepository
	mappings consentdomain.PseudonymousMappingRepository
}

func (a *intentBatchFetcherAdapter) FindLatestByUserIDs(
	ctx context.Context,
	userIDs []int64,
) (map[int64]*adminApp.IntentSummary, error) {
	pseudonymousByUser, err := a.mappings.FindPseudonymousIDsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	if len(pseudonymousByUser) == 0 {
		return map[int64]*adminApp.IntentSummary{}, nil
	}

	ids := make([]string, 0, len(pseudonymousByUser))
	for _, pseudonymousID := range pseudonymousByUser {
		ids = append(ids, pseudonymousID)
	}
	raw, err := a.repo.FindLatestByPseudonymousIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*adminApp.IntentSummary, len(raw))
	for userID, pseudonymousID := range pseudonymousByUser {
		intent, ok := raw[pseudonymousID]
		if !ok {
			continue
		}
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
	mappingRepo := consentInfra.NewPseudonymousMappingRepository(db)

	return adminApp.NewGetUsersWithIntentsUseCase(
		&userListerAdapter{repo: userRepo},
		&intentBatchFetcherAdapter{repo: intentRepo, mappings: mappingRepo},
		logger,
	)
}
