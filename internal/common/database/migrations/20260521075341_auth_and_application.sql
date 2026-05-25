-- migrate:up

CREATE TABLE developers (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100)   NOT NULL,
    email         VARCHAR(150)   NOT NULL UNIQUE,
    password_hash VARCHAR(255)   NOT NULL,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE applications (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    developer_id  UUID           NOT NULL REFERENCES developers(id) ON DELETE CASCADE,
    app_name      VARCHAR(100)   NOT NULL,
    platform      VARCHAR(50)    NOT NULL, -- e.g., 'flutter', 'android', 'ios', 'web'
    bundle_id     VARCHAR(150)   NOT NULL, -- e.g., 'com.shega.app'
    status        VARCHAR(20)    NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_app_status CHECK (status IN ('active', 'suspended', 'development'))
);

CREATE TABLE api_keys (
    id             UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID           NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    key_value      VARCHAR(255)   NOT NULL UNIQUE, -- The actual public/secret key string (e.g., 'pk_live_...')
    is_active      BOOLEAN        NOT NULL DEFAULT TRUE,
    rate_limit     INT            NOT NULL DEFAULT 60, -- Max requests allowed per minute
    created_at     TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    expires_at     TIMESTAMPTZ,   -- NULL means it never naturally expires unless revoked
    
    CONSTRAINT chk_rate_limit CHECK (rate_limit > 0)
);

-- Crucial Performance Indexes for Middleware Validation Lookup
CREATE INDEX idx_api_keys_value ON api_keys(key_value) WHERE is_active = TRUE;
CREATE INDEX idx_applications_developer ON applications(developer_id);

-- migrate:down

DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS developers;
