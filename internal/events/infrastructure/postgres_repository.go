package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"skykin-platform/internal/events/domain"
	"skykin-platform/internal/events/infrastructure/persistence"

	"gorm.io/gorm"
)

type PostgresEventRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresEventRepository {
	return &PostgresEventRepository{db: db}
}

func (r *PostgresEventRepository) SaveBatch(ctx context.Context, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	records := make([]persistence.EventRecord, len(events))
	for i, e := range events {
		meta, err := json.Marshal(e.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
		var appID, sessionID *string
		if e.ApplicationID != "" {
			appID = &e.ApplicationID
		}
		if e.SessionID != "" {
			sessionID = &e.SessionID
		}
		records[i] = persistence.EventRecord{
			EventID:       e.EventID,
			UserID:        e.UserID,
			ApplicationID: appID,
			SessionID:     sessionID,
			EventType:     string(e.EventType),
			Domain:        e.Domain,
			ScreenName:    e.ScreenName,
			Metadata:      meta,
			DeviceType:    e.DeviceType,
			Platform:      e.Platform,
			AppVersion:    e.AppVersion,
			CreatedAt:     e.CreatedAt,
		}
	}

	return r.db.WithContext(ctx).Create(&records).Error
}

func (r *PostgresEventRepository) FindByUser(ctx context.Context, externalUserID string, limit int) ([]domain.Event, error) {
	var records []persistence.EventRecord
	err := r.db.WithContext(ctx).
		Table("events").
		Select("events.*").
		Joins("JOIN users u ON u.id = events.user_id").
		Where("u.external_user_id = ?", externalUserID).
		Order("events.created_at desc").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return toDomainEvents(records)
}

func (r *PostgresEventRepository) FindByUserInternalID(ctx context.Context, internalUserID string, limit int) ([]domain.Event, error) {
	var records []persistence.EventRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", internalUserID).
		Order("created_at desc").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return toDomainEvents(records)
}

func (r *PostgresEventRepository) FindSessionEvents(ctx context.Context, sessionID string, limit int) ([]domain.Event, error) {
	var records []persistence.EventRecord
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at desc").
		Limit(limit).
		Find(&records).Error
	if err != nil {
		return nil, err
	}
	return toDomainEvents(records)
}

func toDomainEvents(records []persistence.EventRecord) ([]domain.Event, error) {
	out := make([]domain.Event, len(records))
	for i, rec := range records {
		var meta map[string]any
		if len(rec.Metadata) > 0 {
			if err := json.Unmarshal(rec.Metadata, &meta); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		appID := ""
		if rec.ApplicationID != nil {
			appID = *rec.ApplicationID
		}
		sessionID := ""
		if rec.SessionID != nil {
			sessionID = *rec.SessionID
		}
		out[i] = domain.Event{
			ID:            rec.ID,
			EventID:       rec.EventID,
			UserID:        rec.UserID,
			ApplicationID: appID,
			EventType:     domain.EventType(rec.EventType),
			Domain:        rec.Domain,
			SessionID:     sessionID,
			ScreenName:    rec.ScreenName,
			Metadata:      meta,
			DeviceType:    rec.DeviceType,
			Platform:      rec.Platform,
			AppVersion:    rec.AppVersion,
			CreatedAt:     rec.CreatedAt,
		}
	}
	return out, nil
}
