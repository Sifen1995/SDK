package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const membershipBatchSize = 1000

type MembershipRepository struct {
	db *gorm.DB
}

func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

var _ domain.MembershipRepository = (*MembershipRepository)(nil)

func (r *MembershipRepository) BulkInsert(ctx context.Context, segmentID uuid.UUID, users []*domain.UserInCandidate) error {
	if len(users) == 0 {
		return nil
	}
	for start := 0; start < len(users); start += membershipBatchSize {
		end := start + membershipBatchSize
		if end > len(users) {
			end = len(users)
		}
		if err := r.insertBatch(ctx, segmentID, users[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func (r *MembershipRepository) insertBatch(ctx context.Context, segmentID uuid.UUID, users []*domain.UserInCandidate) error {
	var b strings.Builder
	b.WriteString(`INSERT INTO segment_memberships (segment_id, pseudonymous_id, confidence, days_active) VALUES `)
	args := make([]interface{}, 0, len(users)*4)
	for i, u := range users {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("(?, ?, ?, ?)")
		args = append(args, segmentID, u.PseudonymousID, u.Confidence, u.DaysActive)
	}
	b.WriteString(` ON CONFLICT (segment_id, pseudonymous_id) DO UPDATE SET confidence = EXCLUDED.confidence, days_active = EXCLUDED.days_active, added_at = CURRENT_TIMESTAMP`)
	return r.db.WithContext(ctx).Exec(b.String(), args...).Error
}

func (r *MembershipRepository) CountMembers(ctx context.Context, segmentID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("segment_memberships").
		Where("segment_id = ?", segmentID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count segment members: %w", err)
	}
	return int(count), nil
}

func (r *MembershipRepository) FindPseudonymousIDsInSegment(ctx context.Context, segmentID uuid.UUID) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT pseudonymous_id FROM segment_memberships
		WHERE segment_id = ?
	`, segmentID).Scan(&ids).Error
	if err != nil {
		return nil, fmt.Errorf("find pseudonymous ids in segment: %w", err)
	}
	return ids, nil
}
