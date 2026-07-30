package infrastructure

import (
	"context"
	"strings"
	"time"

	"skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type candidateRow struct {
	ID                 uuid.UUID
	IntentName         string
	UserCount          int
	AvgConfidence      float64
	AvgDaysActive      float64
	MinDaysActive      int
	LookbackDays       int
	Status             string
	ScannedAt          time.Time
	ReviewedBy         *uuid.UUID
	ReviewedAt         *time.Time
	ReviewNotes        *string
	PublishedSegmentID *uuid.UUID
}

const candidateSelectColumns = `
	id, intent_name, user_count, avg_confidence, avg_days_active,
	min_days_active, lookback_days, status, scanned_at,
	reviewed_by, reviewed_at, review_notes, published_segment_id`

// CandidateRepository persists segment candidates and their captured member lists.
type CandidateRepository struct {
	db *gorm.DB
}

func NewCandidateRepository(db *gorm.DB) *CandidateRepository {
	return &CandidateRepository{db: db}
}

var _ domain.CandidateRepository = (*CandidateRepository)(nil)

// UpsertPending relies on the partial unique index on (intent_name) WHERE status =
// 'pending', so two concurrent scans of the same intent can never both insert.
func (r *CandidateRepository) UpsertPending(
	ctx context.Context,
	c *domain.SegmentCandidate,
	users []*domain.UserInCandidate,
) (domain.UpsertOutcome, error) {
	var outcome domain.UpsertOutcome
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var result struct {
			ID        uuid.UUID
			ScannedAt time.Time
		}
		err := tx.Raw(`
			INSERT INTO segment_candidates (
				id, intent_name, user_count, avg_confidence, avg_days_active,
				min_days_active, lookback_days, status, scanned_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?)
			ON CONFLICT (intent_name) WHERE status = 'pending'
			DO UPDATE SET
				user_count      = EXCLUDED.user_count,
				avg_confidence  = EXCLUDED.avg_confidence,
				avg_days_active = EXCLUDED.avg_days_active,
				min_days_active = EXCLUDED.min_days_active,
				lookback_days   = EXCLUDED.lookback_days,
				scanned_at      = EXCLUDED.scanned_at
			RETURNING id, scanned_at
		`, c.ID, c.IntentName, c.UserCount, c.AvgConfidence, c.AvgDaysActive,
			c.MinDaysActive, c.LookbackDays, c.ScannedAt).Scan(&result).Error
		if err != nil {
			return err
		}
		outcome = domain.UpsertOutcome{CandidateID: result.ID, Created: result.ID == c.ID}
		return replaceCandidateUsers(ctx, tx, result.ID, users)
	})
	if err != nil {
		return domain.UpsertOutcome{}, err
	}
	return outcome, nil
}

func replaceCandidateUsers(
	ctx context.Context,
	tx *gorm.DB,
	candidateID uuid.UUID,
	users []*domain.UserInCandidate,
) error {
	if err := tx.WithContext(ctx).
		Exec(`DELETE FROM segment_candidate_users WHERE candidate_id = ?`, candidateID).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(`INSERT INTO segment_candidate_users (candidate_id, pseudonymous_id, confidence, days_active, last_seen_at) VALUES `)
	args := make([]interface{}, 0, len(users)*5)
	for i, u := range users {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("(?, ?, ?, ?, ?)")
		lastSeen := u.LastSeenAt
		if lastSeen.IsZero() {
			lastSeen = time.Now().UTC()
		}
		args = append(args, candidateID, u.PseudonymousID, u.Confidence, u.DaysActive, lastSeen)
	}
	b.WriteString(` ON CONFLICT (candidate_id, pseudonymous_id) DO UPDATE SET
		confidence = EXCLUDED.confidence,
		days_active = EXCLUDED.days_active,
		last_seen_at = EXCLUDED.last_seen_at`)
	return tx.WithContext(ctx).Exec(b.String(), args...).Error
}

func (r *CandidateRepository) FindByStatus(ctx context.Context, status domain.CandidateStatus) ([]*domain.SegmentCandidate, error) {
	var rows []candidateRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT`+candidateSelectColumns+`
		FROM segment_candidates
		WHERE status = ?
		ORDER BY user_count DESC
	`, string(status)).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapCandidateRows(rows), nil
}

func (r *CandidateRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.SegmentCandidate, error) {
	var row candidateRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT`+candidateSelectColumns+`
		FROM segment_candidates
		WHERE id = ?
	`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	return mapCandidateRow(&row), nil
}

func (r *CandidateRepository) LockPending(ctx context.Context, id uuid.UUID) (*domain.SegmentCandidate, error) {
	var row candidateRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT`+candidateSelectColumns+`
		FROM segment_candidates
		WHERE id = ? AND status = 'pending'
		FOR UPDATE
	`, id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == uuid.Nil {
		return nil, gorm.ErrRecordNotFound
	}
	return mapCandidateRow(&row), nil
}

func (r *CandidateRepository) GetUsers(ctx context.Context, candidateID uuid.UUID) ([]*domain.UserInCandidate, error) {
	var rows []struct {
		PseudonymousID string
		Confidence     float64
		DaysActive     int
		LastSeenAt     time.Time
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT pseudonymous_id, confidence, days_active, last_seen_at
		FROM segment_candidate_users
		WHERE candidate_id = ?
		ORDER BY confidence DESC
	`, candidateID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]*domain.UserInCandidate, 0, len(rows))
	for i := range rows {
		out = append(out, &domain.UserInCandidate{
			PseudonymousID: rows[i].PseudonymousID,
			Confidence:     rows[i].Confidence,
			DaysActive:     rows[i].DaysActive,
			LastSeenAt:     rows[i].LastSeenAt,
		})
	}
	return out, nil
}

func (r *CandidateRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CandidateStatus, reviewedBy uuid.UUID, notes string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Exec(`
		UPDATE segment_candidates
		SET status = ?, reviewed_by = ?, reviewed_at = ?, review_notes = ?
		WHERE id = ?
	`, string(status), reviewedBy, now, notes, id).Error
}

func (r *CandidateRepository) LinkToSegment(ctx context.Context, candidateID, segmentID uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE segment_candidates
		SET published_segment_id = ?
		WHERE id = ?
	`, segmentID, candidateID).Error
}

func mapCandidateRows(rows []candidateRow) []*domain.SegmentCandidate {
	out := make([]*domain.SegmentCandidate, 0, len(rows))
	for i := range rows {
		out = append(out, mapCandidateRow(&rows[i]))
	}
	return out
}

func mapCandidateRow(row *candidateRow) *domain.SegmentCandidate {
	notes := ""
	if row.ReviewNotes != nil {
		notes = *row.ReviewNotes
	}
	return &domain.SegmentCandidate{
		ID: row.ID, IntentName: row.IntentName, UserCount: row.UserCount,
		AvgConfidence: row.AvgConfidence, AvgDaysActive: row.AvgDaysActive,
		MinDaysActive: row.MinDaysActive, LookbackDays: row.LookbackDays,
		Status: domain.CandidateStatus(row.Status), ScannedAt: row.ScannedAt,
		ReviewedBy: row.ReviewedBy, ReviewedAt: row.ReviewedAt, ReviewNotes: notes,
		PublishedSegmentID: row.PublishedSegmentID,
	}
}
