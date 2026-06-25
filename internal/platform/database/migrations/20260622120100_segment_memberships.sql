-- Pre-computed segment membership from approved classification candidates.
CREATE TABLE IF NOT EXISTS segment_memberships (
    segment_id   UUID NOT NULL REFERENCES audience_segments(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    confidence   NUMERIC(5,3) NOT NULL,
    days_active  INTEGER      NOT NULL,
    added_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (segment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_seg_memberships_segment ON segment_memberships (segment_id);
CREATE INDEX IF NOT EXISTS idx_seg_memberships_user    ON segment_memberships (user_id);
