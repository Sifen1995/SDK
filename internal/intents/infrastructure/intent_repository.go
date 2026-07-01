package infrastructure

import (
	"context"
	"time"

	"skykin-platform/configs"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	intentdomain "skykin-platform/internal/intents/domain"
	"skykin-platform/internal/intents/infrastructure/persistence"

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

func (r *intentRepository) Create(ctx context.Context, intent *intentdomain.Intent) (*intentdomain.Intent, error) {
	row := persistence.IntentRowFromDomain(intent)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return row.ToDomain(), nil
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

// ConsistencyReader exposes sustained-intent queries for segment classification.
type ConsistencyReader struct {
	inner *intentRepository
}

func NewConsistencyReader(db *gorm.DB, cfg *configs.Config) *ConsistencyReader {
	return &ConsistencyReader{inner: &intentRepository{db: db, config: cfg}}
}

func (r *ConsistencyReader) FindConsistentUsers(
	ctx context.Context,
	intentName string,
	minConf float64,
	lookbackDays int,
	minDays int,
	maxAgeDays int,
) ([]*analyticsdomain.ConsistentUser, error) {
	return r.inner.findConsistentUsers(ctx, intentName, minConf, lookbackDays, minDays, maxAgeDays)
}

func (r *intentRepository) findConsistentUsers(
	ctx context.Context,
	intentName string,
	minConf float64,
	lookbackDays int,
	minDays int,
	maxAgeDays int,
) ([]*analyticsdomain.ConsistentUser, error) {
	type row struct {
		UserID        uuid.UUID
		DaysActive    int
		AvgConfidence float64
		LastSeenAt    time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			user_id,
			COUNT(DISTINCT DATE(created_at))   AS days_active,
			ROUND(AVG(confidence)::numeric, 3) AS avg_confidence,
			MAX(created_at)                    AS last_seen_at
		FROM intents
		WHERE intent_name  = ?
		AND   confidence  >= ?
		AND   created_at  >= NOW() - make_interval(days => ?::int)
		GROUP BY user_id
		HAVING
			COUNT(DISTINCT DATE(created_at)) >= ?
			AND MAX(created_at) >= NOW() - make_interval(days => ?::int)
		ORDER BY days_active DESC, avg_confidence DESC
	`, intentName, minConf, lookbackDays, minDays, maxAgeDays).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*analyticsdomain.ConsistentUser, 0, len(rows))
	for i := range rows {
		out = append(out, &analyticsdomain.ConsistentUser{
			UserID:     rows[i].UserID,
			Confidence: rows[i].AvgConfidence,
			DaysActive: rows[i].DaysActive,
			LastSeenAt: rows[i].LastSeenAt,
		})
	}
	return out, nil
}
