-- migrate:up

CREATE TABLE IF NOT EXISTS sdk_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id        VARCHAR(255) NOT NULL UNIQUE,
    user_id         VARCHAR(255) NOT NULL,
    application_id  UUID,
    event_type      VARCHAR(100) NOT NULL,
    domain          VARCHAR(100),
    session_id      VARCHAR(255),
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at     TIMESTAMPTZ NOT NULL,
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_flagged      BOOLEAN NOT NULL DEFAULT FALSE,
    flag_reason     VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_sdk_events_user_id ON sdk_events (user_id);
CREATE INDEX IF NOT EXISTS idx_sdk_events_session_id ON sdk_events (session_id);
CREATE INDEX IF NOT EXISTS idx_sdk_events_event_type ON sdk_events (event_type);
CREATE INDEX IF NOT EXISTS idx_sdk_events_application_id ON sdk_events (application_id);
CREATE INDEX IF NOT EXISTS idx_sdk_events_occurred_at ON sdk_events (occurred_at DESC);

-- migrate:down

DROP TABLE IF EXISTS sdk_events;
