package infrastructure

import (
	"context"

	deliverydomain "skykin-platform/internal/delivery/domain"
	"skykin-platform/internal/delivery/infrastructure/persistence"

	"gorm.io/gorm"
)

type DemoSMSRecipientRepository struct {
	db *gorm.DB
}

func NewDemoSMSRecipientRepository(db *gorm.DB) *DemoSMSRecipientRepository {
	return &DemoSMSRecipientRepository{db: db}
}

var _ deliverydomain.DemoSMSRecipientRepository = (*DemoSMSRecipientRepository)(nil)

func (r *DemoSMSRecipientRepository) FindActiveByPseudonymousID(
	ctx context.Context,
	pseudonymousID string,
) (*deliverydomain.DemoSMSRecipient, error) {
	var row persistence.DemoSMSRecipientRow
	err := r.db.WithContext(ctx).
		Where("pseudonymous_id = ? AND is_active = ?", pseudonymousID, true).
		First(&row).Error
	if err == gorm.ErrRecordNotFound {
		// Fallback for rows not yet backfilled with denormalized pseudonymous_id.
		err = r.db.WithContext(ctx).
			Table("demo_sms_recipients AS rec").
			Select("rec.*").
			Joins("JOIN pseudonymous_mappings pm ON pm.user_id = rec.user_id").
			Where("pm.pseudonymous_id = ? AND rec.is_active = ?", pseudonymousID, true).
			First(&row).Error
	}
	if err != nil {
		return nil, err
	}
	return &deliverydomain.DemoSMSRecipient{
		UserID:             row.UserID,
		PhoneE164:          row.PhoneE164,
		DisplayName:        row.DisplayName,
		IsActive:           row.IsActive,
		IsMock:             row.IsMock,
		ProviderExternalID: row.ProviderExternalID,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}
