-- migrate:up
-- Production behavioral events schema (domain-agnostic).
-- Migrates existing sdk_events rows without dropping data.

CREATE TABLE IF NOT EXISTS events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        UUID NOT NULL UNIQUE,
    user_id         UUID NOT NULL,
    application_id  UUID,
    session_id      UUID,
    event_type      VARCHAR(100) NOT NULL,
    domain          VARCHAR(100),
    screen_name     VARCHAR(255),
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    device_type     VARCHAR(50),
    platform        VARCHAR(50),
    app_version     VARCHAR(50),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_user_id ON events (user_id);
CREATE INDEX IF NOT EXISTS idx_events_session_id ON events (session_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events (event_type);
CREATE INDEX IF NOT EXISTS idx_events_domain ON events (domain);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_metadata_gin ON events USING GIN (metadata);

-- Backfill from legacy sdk_events when present.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'sdk_events'
    ) THEN
        INSERT INTO events (
            event_id,
            user_id,
            application_id,
            session_id,
            event_type,
            domain,
            screen_name,
            metadata,
            device_type,
            platform,
            app_version,
            created_at
        )
        SELECT
            CASE
                WHEN se.event_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
                    THEN se.event_id::uuid
                ELSE gen_random_uuid()
            END,
            COALESCE(u.id, gen_random_uuid()),
            se.application_id,
            CASE
                WHEN se.session_id IS NOT NULL
                     AND se.session_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
                    THEN se.session_id::uuid
                ELSE NULL
            END,
            se.event_type,
            se.domain,
            COALESCE(se.metadata->>'screen_name', NULL),
            CASE
                WHEN se.event_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
                    THEN se.metadata
                ELSE se.metadata || jsonb_build_object('_legacy_event_id', se.event_id)
            END,
            se.metadata->>'device_type',
            se.metadata->>'platform',
            se.metadata->>'app_version',
            COALESCE(se.occurred_at, se.received_at, NOW())
        FROM sdk_events se
        LEFT JOIN users u ON u.external_user_id = se.user_id
        ON CONFLICT (event_id) DO NOTHING;
    END IF;
END $$;

-- migrate:down
-- Intentionally non-destructive: legacy sdk_events is preserved.
DROP TABLE IF EXISTS events;
