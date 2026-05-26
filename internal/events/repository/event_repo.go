package repository

import (
	"context"
	"skykin-platform/configs"
	"skykin-platform/internal/events/model"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, event *model.Event) error
	CreateBatch(ctx context.Context, events []model.Event) error
	GetRecentEvents(ctx context.Context, userID string, limit int) ([]model.Event, error)
}

type eventRepo struct {
	db     *gorm.DB
	config *configs.Config
}

func NewEventRepository(db *gorm.DB, cfg *configs.Config) EventRepository {
	return &eventRepo{db: db, config: cfg}
}

func (r *eventRepo) Create(ctx context.Context, event *model.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *eventRepo) CreateBatch(ctx context.Context, events []model.Event) error {
	return r.db.WithContext(ctx).Create(&events).Error
}

func (r *eventRepo) GetRecentEvents(ctx context.Context, userID string, limit int) ([]model.Event, error) {
	var events []model.Event
	// Get the last N events ordered by timestamp descending
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("timestamp desc").
		Limit(limit).
		Find(&events).Error
	return events, err
}
