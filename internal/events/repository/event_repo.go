package repository

import (
	"context"
	"skykin-platform/configs"
	"skykin-platform/internal/events/model"

	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, event *model.Event) error
	GetRecentEvents(ctx context.Context, userID string, limit int) ([]model.Event, error)
}

type eventRepo struct {
	db     *gorm.DB
	config *configs.Config
}

func NewEventRepository(db *gorm.DB, cfg *configs.Config) EventRepository {
	return &eventRepo{db: db, config: cfg}
}
