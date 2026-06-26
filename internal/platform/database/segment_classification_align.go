package database

import (
	"log"

	"gorm.io/gorm"
)

func alignSegmentClassificationSchema(db *gorm.DB) {
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
		`CREATE TABLE IF NOT EXISTS segment_memberships (
			segment_id   UUID NOT NULL REFERENCES audience_segments(id) ON DELETE CASCADE,
			user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			confidence   NUMERIC(5,3) NOT NULL,
			days_active  INTEGER      NOT NULL,
			added_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			PRIMARY KEY (segment_id, user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_memberships_segment ON segment_memberships (segment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_seg_memberships_user    ON segment_memberships (user_id)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("segment classification schema align (non-fatal): %v", err)
		}
	}
	log.Println("segment classification schema aligned")
}
