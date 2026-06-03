package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"skykin-platform/internal/campaigns/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, c *model.Campaign) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) Get(ctx context.Context, id string) (*model.Campaign, error) {
	var c model.Campaign
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) ListByAdvertiser(ctx context.Context, advertiserID string) ([]model.Campaign, error) {
	var list []model.Campaign
	err := r.db.WithContext(ctx).Where("advertiser_id = ?", advertiserID).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) ListAll(ctx context.Context) ([]model.Campaign, error) {
	var list []model.Campaign
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *Repository) Update(ctx context.Context, c *model.Campaign) error {
	return r.db.WithContext(ctx).Save(c).Error
}

// FindActiveForIntent returns the newest active campaign matching intent, app, and format.
func (r *Repository) FindActiveForIntent(ctx context.Context, targetIntent, applicationID, creativeFormat string) (*model.Campaign, error) {
	var c model.Campaign
	q := r.db.WithContext(ctx).
		Where("target_intent = ? AND is_active = ? AND validation_status = ?", targetIntent, true, "passed")
	if applicationID != "" {
		q = q.Where("application_id = ?", applicationID)
	}
	if creativeFormat != "" {
		q = q.Where("creative_format = ?", creativeFormat)
	}
	if err := q.Order("created_at desc").First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) LogDelivery(ctx context.Context, log *model.DeliveryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func ParseCanvasJSON(raw []byte) (map[string]any, error) {
	var m map[string]any
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// CampaignAdContent builds SDK/WebSocket payload from a campaign row.
func CampaignAdContent(c *model.Campaign) (map[string]any, error) {
	canvas, err := ParseCanvasJSON(c.CanvasJSON)
	if err != nil {
		return nil, fmt.Errorf("canvas_json: %w", err)
	}
	content := map[string]any{
		"title":            c.Title,
		"body_text":        c.BodyText,
		"image_url":        c.ImageURL,
		"destination_url":  c.DestinationURL,
		"creative_format":  c.CreativeFormat,
		"canvas_json":      canvas,
	}
	return content, nil
}
