package infrastructure

import (
	"context"
	"sync"
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

// CandidateRepository persists segment candidates and holds user lists in memory.
type CandidateRepository struct {
	db    *gorm.DB
	users sync.Map
}

func NewCandidateRepository(db *gorm.DB) *CandidateRepository {
	return &CandidateRepository{db: db}
}

var _ domain.CandidateRepository = (*CandidateRepository)(nil)

func (r *CandidateRepository) Save(ctx context.Context, c *domain.SegmentCandidate, users []*domain.UserInCandidate) error {
	notes := c.ReviewNotes
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO segment_candidates (
			id, intent_name, user_count, avg_confidence, avg_days_active,
			min_days_active, lookback_days, status, scanned_at,
			reviewed_by, reviewed_at, review_notes, published_segment_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, c.ID, c.IntentName, c.UserCount, c.AvgConfidence, c.AvgDaysActive,
		c.MinDaysActive, c.LookbackDays, string(c.Status), c.ScannedAt,
		c.ReviewedBy, c.ReviewedAt, notesPtr, c.PublishedSegmentID).Error
	if err != nil {
		return err
	}
	r.users.Store(c.ID.String(), users)
	return nil
}

func (r *CandidateRepository) FindByStatus(ctx context.Context, status domain.CandidateStatus) ([]*domain.SegmentCandidate, error) {
	var rows []candidateRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, intent_name, user_count, avg_confidence, avg_days_active,
		       min_days_active, lookback_days, status, scanned_at,
		       reviewed_by, reviewed_at, review_notes, published_segment_id
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
		SELECT id, intent_name, user_count, avg_confidence, avg_days_active,
		       min_days_active, lookback_days, status, scanned_at,
		       reviewed_by, reviewed_at, review_notes, published_segment_id
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

func (r *CandidateRepository) GetUsers(ctx context.Context, candidateID uuid.UUID) ([]*domain.UserInCandidate, error) {
	_ = ctx
	val, ok := r.users.Load(candidateID.String())
	if !ok {
		return []*domain.UserInCandidate{}, nil
	}
	users, _ := val.([]*domain.UserInCandidate)
	return users, nil
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
