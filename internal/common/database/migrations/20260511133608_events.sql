-- migrate:up

CREATE TABLE events (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type   VARCHAR(50) NOT NULL,
    metadata     JSONB        NOT NULL DEFAULT '{}',
    timestamp    TIMESTAMPTZ  NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_event_type CHECK (event_type IN ('search', 'product_view', 'category_view', 'add_to_cart', 'remove_from_cart', 'signup_complete', 'checkout_started'))
);
CREATE INDEX idx_events_user_timestamp ON events (user_id, timestamp DESC);
CREATE INDEX idx_events_metadata_gin ON events USING GIN (metadata);

-- migrate:down
DROP TABLE IF EXISTS events;


