package persistance

import (
	"context"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"gorm.io/gorm"
)

// SyncRepository serves the mobile fraud-rule cache from PostgreSQL.
type SyncRepository struct {
	db *gorm.DB
}

func NewSyncRepository(db *gorm.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

var _ frauddomain.SyncRepository = (*SyncRepository)(nil)

func (r *SyncRepository) Sync(
	ctx context.Context,
	since *time.Time,
	until time.Time,
) (*frauddomain.SyncSnapshot, error) {
	snapshot := &frauddomain.SyncSnapshot{
		BlockedDomains: make([]frauddomain.BlockedDomain, 0),
		BlockedSenders: make([]frauddomain.BlockedSender, 0),
		ScamPatterns:   make([]frauddomain.ScamPattern, 0),
		NextCursor:     until,
		IsDelta:        since != nil,
	}

	var domainRows []BlockedDomainsRow
	domainQuery := r.db.WithContext(ctx).
		Where("updated_at <= ?", until)
	if since == nil {
		domainQuery = domainQuery.
			Where("status = ?", frauddomain.StatusActive).
			Where("(expires_at IS NULL OR expires_at > ?)", until)
	} else {
		// Expiration is also a state transition. Include domains whose expiry
		// crossed the cursor window even if no database UPDATE occurred.
		domainQuery = domainQuery.Where(
			"(updated_at > ? OR (expires_at > ? AND expires_at <= ?))",
			*since,
			*since,
			until,
		)
	}
	if err := domainQuery.
		Order("updated_at ASC").
		Order("domain ASC").
		Find(&domainRows).Error; err != nil {
		return nil, err
	}
	for _, row := range domainRows {
		value := row.ToDomain()
		if since != nil && value.ExpiresAt != nil && !value.ExpiresAt.After(until) {
			// Mobile clients can apply one tombstone rule for explicit revocation
			// and natural expiry.
			value.Status = frauddomain.StatusRevoked
		}
		snapshot.BlockedDomains = append(snapshot.BlockedDomains, value)
	}

	var senderRows []BlockedSendersRow
	senderQuery := r.db.WithContext(ctx).
		Where("updated_at <= ?", until)
	if since == nil {
		senderQuery = senderQuery.Where("status = ?", frauddomain.StatusActive)
	} else {
		senderQuery = senderQuery.Where("updated_at > ?", *since)
	}
	if err := senderQuery.
		Order("updated_at ASC").
		Order("sender_hash ASC").
		Find(&senderRows).Error; err != nil {
		return nil, err
	}
	for _, row := range senderRows {
		snapshot.BlockedSenders = append(snapshot.BlockedSenders, row.ToDomain())
	}

	var patternRows []ScamPatternsRow
	patternQuery := r.db.WithContext(ctx).
		Where("updated_at <= ?", until)
	if since == nil {
		patternQuery = patternQuery.Where("is_active = ?", true)
	} else {
		patternQuery = patternQuery.Where("updated_at > ?", *since)
	}
	if err := patternQuery.
		Order("updated_at ASC").
		Order("id ASC").
		Find(&patternRows).Error; err != nil {
		return nil, err
	}
	for _, row := range patternRows {
		snapshot.ScamPatterns = append(snapshot.ScamPatterns, row.ToDomain())
	}

	return snapshot, nil
}
