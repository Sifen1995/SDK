-- Analyst profiles for read_only_analyst portal users (separate from advertisers).
CREATE TABLE IF NOT EXISTS analysts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    display_name VARCHAR(255) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE portal_users
    ADD COLUMN IF NOT EXISTS analyst_id UUID REFERENCES analysts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_portal_users_analyst
    ON portal_users (analyst_id);
