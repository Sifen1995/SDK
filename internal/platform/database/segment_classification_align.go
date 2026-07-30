package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// alignSegmentClassificationSchema owns the DDL for segment candidates, their
// captured user lists and published memberships. Failures are fatal: booting with
// a missing or invalid membership table lets candidate approval publish empty
// segments, which is worse than refusing to start.
func alignSegmentClassificationSchema(db *gorm.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS segment_candidates (
			id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
			intent_name          VARCHAR(100) NOT NULL,
			user_count           INTEGER      NOT NULL DEFAULT 0,
			avg_confidence       NUMERIC(5,3) NOT NULL,
			avg_days_active      NUMERIC(5,1) NOT NULL,
			min_days_active      INTEGER      NOT NULL,
			lookback_days        INTEGER      NOT NULL,
			status               VARCHAR(20)  NOT NULL DEFAULT 'pending',
			scanned_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			reviewed_by          UUID,
			reviewed_at          TIMESTAMPTZ,
			review_notes         TEXT,
			published_segment_id UUID REFERENCES audience_segments(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_candidates_status ON segment_candidates (status)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_candidates_intent ON segment_candidates (intent_name)`,

		// Candidate user lists used to live in process memory, so they were lost on
		// restart and invisible to other instances.
		`CREATE TABLE IF NOT EXISTS segment_candidate_users (
			candidate_id    UUID         NOT NULL REFERENCES segment_candidates(id) ON DELETE CASCADE,
			pseudonymous_id UUID         NOT NULL,
			confidence      NUMERIC(5,3) NOT NULL,
			days_active     INTEGER      NOT NULL,
			last_seen_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			PRIMARY KEY (candidate_id, pseudonymous_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_candidate_users_candidate ON segment_candidate_users (candidate_id)`,

		// Membership is keyed by pseudonymous id; there is deliberately no foreign
		// key to users because the platform never joins behaviour back to users.id.
		`CREATE TABLE IF NOT EXISTS segment_memberships (
			segment_id      UUID NOT NULL REFERENCES audience_segments(id) ON DELETE CASCADE,
			pseudonymous_id UUID NOT NULL,
			confidence      NUMERIC(5,3) NOT NULL,
			days_active     INTEGER      NOT NULL,
			added_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			PRIMARY KEY (segment_id, pseudonymous_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_memberships_segment         ON segment_memberships (segment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_memberships_pseudonymous_id ON segment_memberships (pseudonymous_id)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("align segment classification schema: %w", err)
		}
	}

	if err := enforceSinglePendingCandidatePerIntent(db); err != nil {
		return err
	}

	log.Println("segment classification schema aligned")
	return nil
}

// enforceSinglePendingCandidatePerIntent is a one-time schema bootstrap helper:
// collapse legacy duplicate pending candidates, then create the unique partial
// index that makes the invariant structural. Future scans rely on UpsertPending.
func enforceSinglePendingCandidatePerIntent(db *gorm.DB) error {
	if err := db.Exec(`
		UPDATE segment_candidates c
		SET status = 'superseded',
		    review_notes = COALESCE(c.review_notes, 'superseded by a newer pending scan')
		WHERE c.status = 'pending'
		  AND EXISTS (
		      SELECT 1 FROM segment_candidates newer
		      WHERE newer.intent_name = c.intent_name
		        AND newer.status = 'pending'
		        AND (newer.scanned_at, newer.id) > (c.scanned_at, c.id)
		  )
	`).Error; err != nil {
		return fmt.Errorf("collapse duplicate pending candidates: %w", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uq_seg_candidates_pending_intent
			ON segment_candidates (intent_name) WHERE status = 'pending'
	`).Error; err != nil {
		return fmt.Errorf("create pending candidate unique index: %w", err)
	}
	return nil
}
