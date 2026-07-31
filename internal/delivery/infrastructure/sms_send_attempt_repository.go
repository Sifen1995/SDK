package infrastructure

import (
	"context"

	deliverydomain "skykin-platform/internal/delivery/domain"
	"skykin-platform/internal/delivery/infrastructure/persistence"

	"gorm.io/gorm"
)

type SMSSendAttemptRepository struct {
	db *gorm.DB
}

func NewSMSSendAttemptRepository(db *gorm.DB) *SMSSendAttemptRepository {
	return &SMSSendAttemptRepository{db: db}
}

var _ deliverydomain.SMSSendAttemptRepository = (*SMSSendAttemptRepository)(nil)

func (r *SMSSendAttemptRepository) Create(ctx context.Context, attempt *deliverydomain.SMSSendAttempt) error {
	row := persistence.SMSSendAttemptRow{
		ID:                attempt.ID,
		SendKey:           attempt.SendKey,
		CampaignID:        attempt.CampaignID,
		PseudonymousID:    attempt.PseudonymousID,
		UserID:            attempt.UserID,
		PhoneE164:         attempt.PhoneE164,
		Provider:          attempt.Provider,
		ProviderMessageID: attempt.ProviderMessageID,
		Status:            attempt.Status,
		MessageBody:       attempt.MessageBody,
		TrackingToken:     attempt.TrackingToken,
		DestinationURL:    attempt.DestinationURL,
		ErrorMessage:      attempt.ErrorMessage,
		SentAt:            attempt.SentAt,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	attempt.ID = row.ID
	attempt.CreatedAt = row.CreatedAt
	attempt.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *SMSSendAttemptRepository) ExistsBySendKey(ctx context.Context, sendKey string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&persistence.SMSSendAttemptRow{}).
		Where("send_key = ?", sendKey).
		Count(&count).Error
	return count > 0, err
}

func (r *SMSSendAttemptRepository) FindByTrackingToken(
	ctx context.Context,
	trackingToken string,
) (*deliverydomain.SMSSendAttempt, error) {
	var row persistence.SMSSendAttemptRow
	if err := r.db.WithContext(ctx).
		Where("tracking_token = ?", trackingToken).
		First(&row).Error; err != nil {
		return nil, err
	}
	return mapSMSSendAttemptRow(&row), nil
}

func (r *SMSSendAttemptRepository) FindBySendKey(
	ctx context.Context,
	sendKey string,
) (*deliverydomain.SMSSendAttempt, error) {
	var row persistence.SMSSendAttemptRow
	if err := r.db.WithContext(ctx).
		Where("send_key = ?", sendKey).
		First(&row).Error; err != nil {
		return nil, err
	}
	return mapSMSSendAttemptRow(&row), nil
}

func (r *SMSSendAttemptRepository) UpdateStatus(
	ctx context.Context,
	attemptID, status, providerMessageID, errorMessage string,
) error {
	updates := map[string]any{
		"status":              status,
		"provider_message_id": providerMessageID,
		"error_message":       errorMessage,
	}
	if status == deliverydomain.SMSSendStatusSent {
		updates["sent_at"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).
		Model(&persistence.SMSSendAttemptRow{}).
		Where("id = ?", attemptID).
		Updates(updates).Error
}

func (r *SMSSendAttemptRepository) ListRecent(ctx context.Context, limit int) ([]deliverydomain.SMSSendAttempt, error) {
	if limit <= 0 {
		limit = 20
	}
	type recentRow struct {
		persistence.SMSSendAttemptRow
		CampaignName string `gorm:"column:campaign_name"`
		ImageURL     string `gorm:"column:image_url"`
	}
	var rows []recentRow
	if err := r.db.WithContext(ctx).
		Table("sms_send_attempts AS a").
		Select(`a.*, c.name AS campaign_name, COALESCE(c.image_url, '') AS image_url`).
		Joins("LEFT JOIN campaigns c ON c.id = a.campaign_id").
		Order("a.created_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]deliverydomain.SMSSendAttempt, 0, len(rows))
	for i := range rows {
		attempt := mapSMSSendAttemptRow(&rows[i].SMSSendAttemptRow)
		attempt.CampaignName = rows[i].CampaignName
		attempt.ImageURL = rows[i].ImageURL
		out = append(out, *attempt)
	}
	return out, nil
}

func mapSMSSendAttemptRow(row *persistence.SMSSendAttemptRow) *deliverydomain.SMSSendAttempt {
	if row == nil {
		return nil
	}
	return &deliverydomain.SMSSendAttempt{
		ID:                row.ID,
		SendKey:           row.SendKey,
		CampaignID:        row.CampaignID,
		PseudonymousID:    row.PseudonymousID,
		UserID:            row.UserID,
		PhoneE164:         row.PhoneE164,
		Provider:          row.Provider,
		ProviderMessageID: row.ProviderMessageID,
		Status:            row.Status,
		MessageBody:       row.MessageBody,
		TrackingToken:     row.TrackingToken,
		DestinationURL:    row.DestinationURL,
		ErrorMessage:      row.ErrorMessage,
		SentAt:            row.SentAt,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}
