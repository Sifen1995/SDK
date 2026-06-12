package infrastructure

import (
	"context"
	"time"

	"skykin-platform/configs"
	intentdomain "skykin-platform/internal/intents/domain"
	"skykin-platform/internal/intents/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type intentRepository struct {
	db     *gorm.DB
	config *configs.Config
}

func NewIntentRepository(db *gorm.DB, cfg *configs.Config) intentdomain.IntentRepository {
	return &intentRepository{db: db, config: cfg}
}

func (r *intentRepository) Create(ctx context.Context, intent *model.Intent) (*model.Intent, error) {
	if err := r.db.WithContext(ctx).Create(intent).Error; err != nil {
		return nil, err
	}
	return intent, nil
}

func (r *intentRepository) FindUsersWithIntent(
	ctx context.Context,
	intentName string,
	minConfidence float64,
	since time.Time,
) ([]uuid.UUID, error) {
	var userIDs []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (user_id) user_id::text
		FROM intents
		WHERE intent_name = ? AND confidence >= ? AND created_at >= ?
		ORDER BY user_id, created_at DESC
	`, intentName, minConfidence, since).Scan(&userIDs).Error
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		out = append(out, parsed)
	}
	return out, nil
}

// FindUsersWithAnyIntent returns distinct users with a recent prediction for any listed intent.
func (r *intentRepository) FindUsersWithAnyIntent(
	ctx context.Context,
	intentNames []string,
	minConfidence float64,
	since time.Time,
) ([]uuid.UUID, error) {
	if len(intentNames) == 0 {
		return nil, nil
	}
	var userIDs []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (user_id) user_id::text
		FROM intents
		WHERE intent_name IN ? AND confidence >= ? AND created_at >= ?
		ORDER BY user_id, created_at DESC
	`, intentNames, minConfidence, since).Scan(&userIDs).Error
	if err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(userIDs))
	for _, id := range userIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		out = append(out, parsed)
	}
	return out, nil
}
