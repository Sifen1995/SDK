package infrastructure

import (
	"context"
	"time"

	"skykin-platform/configs"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	intentdomain "skykin-platform/internal/intents/domain"
	"skykin-platform/internal/intents/infrastructure/persistence"

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
) ([]string, error) {
	return r.findDistinctUsersWithFilter(ctx, func(q *gorm.DB) *gorm.DB {
		return q.Where("intent_name = ? AND confidence >= ? AND created_at >= ?", intentName, minConfidence, since)
	})
}

func (r *intentRepository) FindUsersWithAnyIntent(
	ctx context.Context,
	intentNames []string,
	minConfidence float64,
	since time.Time,
) ([]string, error) {
	if len(intentNames) == 0 {
		return nil, nil
	}
	return r.findDistinctUsersWithFilter(ctx, func(q *gorm.DB) *gorm.DB {
		return q.Where("intent_name IN ? AND confidence >= ? AND created_at >= ?", intentNames, minConfidence, since)
	})
}

func (r *intentRepository) findDistinctUsersWithFilter(
	ctx context.Context,
	apply func(*gorm.DB) *gorm.DB,
) ([]string, error) {
	filtered := apply(r.db.WithContext(ctx).Model(&persistence.IntentRow{}))
	sub := filtered.
		Select("user_id, MAX(created_at) AS created_at").
		Group("user_id")

	var rows []struct {
		UserID string
	}
	err := r.db.WithContext(ctx).
		Table("intents").
		Select("intents.user_id").
		Joins("INNER JOIN (?) AS latest ON intents.user_id = latest.user_id AND intents.created_at = latest.created_at", sub).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.UserID != "" {
			out = append(out, row.UserID)
		}
	}
	return out, nil
}

func (r *intentRepository) FindLatestByUserIDs(
	ctx context.Context,
	userIDs []string,
) (map[string]*intentdomain.Intent, error) {
	if len(userIDs) == 0 {
		return map[string]*intentdomain.Intent{}, nil
	}

	sub := r.db.WithContext(ctx).
		Model(&persistence.IntentRow{}).
		Select("user_id, MAX(created_at) AS created_at").
		Where("user_id IN ?", userIDs).
		Group("user_id")

	var rows []persistence.IntentRow
	err := r.db.WithContext(ctx).
		Table("intents").
		Joins("INNER JOIN (?) AS latest ON intents.user_id = latest.user_id AND intents.created_at = latest.created_at", sub).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]*intentdomain.Intent, len(rows))
	for i := range rows {
		d := rows[i].ToDomain()
		result[d.UserID] = d
	}
	return result, nil
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
	lookbackSince := time.Now().AddDate(0, 0, -lookbackDays)
	maxAgeSince := time.Now().AddDate(0, 0, -maxAgeDays)

	type row struct {
		UserID        string
		DaysActive    int
		AvgConfidence float64
		LastSeenAt    time.Time
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&persistence.IntentRow{}).
		Select(`
			user_id,
			COUNT(DISTINCT DATE(created_at)) AS days_active,
			ROUND(AVG(confidence)::numeric, 3) AS avg_confidence,
			MAX(created_at) AS last_seen_at`).
		Where("intent_name = ? AND confidence >= ? AND created_at >= ?", intentName, minConf, lookbackSince).
		Group("user_id").
		Having("COUNT(DISTINCT DATE(created_at)) >= ? AND MAX(created_at) >= ?", minDays, maxAgeSince).
		Order("days_active DESC, avg_confidence DESC").
		Scan(&rows).Error
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
