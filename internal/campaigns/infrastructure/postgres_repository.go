package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	billingdomain "skykin-platform/internal/billing/domain"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/campaigns/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var (
	_ campaigndomain.CampaignRepository = (*Repository)(nil)
	_ billingdomain.CampaignQuotaReader  = (*Repository)(nil)
)

func (r *Repository) ListActive(ctx context.Context) ([]model.Campaign, error) {
	var list []model.Campaign
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND validation_status = ?", true, "passed").
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

func (r *Repository) Create(ctx context.Context, c *model.Campaign) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// CreateTx inserts a campaign inside an existing database transaction.
func (r *Repository) CreateTx(ctx context.Context, tx *gorm.DB, c *model.Campaign) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Transaction runs fn inside a single database transaction.
func (r *Repository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// CountActiveByAdvertiser counts is_active campaigns for subscription quota enforcement.
func (r *Repository) CountActiveByAdvertiser(ctx context.Context, advertiserID string) (int, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.Campaign{}).
		Where("advertiser_id = ? AND is_active = ?", advertiserID, true).
		Count(&n).Error
	return int(n), err
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

// FindActiveForIntent returns the newest active campaign matching intent and channel code.
func (r *Repository) FindActiveForIntent(ctx context.Context, targetIntent, channelCode string) (*model.Campaign, error) {
	var c model.Campaign
	q := r.db.WithContext(ctx).
		Table("campaigns").
		Select("campaigns.*").
		Joins("JOIN channels ON channels.id = campaigns.channel_id").
		Where("campaigns.target_intent = ? AND campaigns.is_active = ? AND campaigns.validation_status = ?",
			targetIntent, true, "passed")
	if channelCode != "" {
		q = q.Where("channels.code = ?", channelCode)
	}
	if err := q.Order("campaigns.created_at desc").First(&c).Error; err != nil {
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
func CampaignAdContent(c *model.Campaign, channelCode string) (map[string]any, error) {
	canvas, err := ParseCanvasJSON(c.CanvasJSON)
	if err != nil {
		return nil, fmt.Errorf("canvas_json: %w", err)
	}
	content := map[string]any{
		"title":           c.Title,
		"body_text":       c.BodyText,
		"image_url":       c.ImageURL,
		"destination_url": c.DestinationURL,
		"channel_code":    channelCode,
		"canvas_json":     canvas,
	}
	return content, nil
}
