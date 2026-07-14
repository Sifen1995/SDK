DROP TABLE IF EXISTS pseudonymous_mappings CASCADE;
DROP TABLE IF EXISTS consents CASCADE;
DROP TABLE IF EXISTS consent_records CASCADE;
DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE users (
    id         BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE consents (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_level    VARCHAR(20)  NOT NULL,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    granted_at       TIMESTAMPTZ,
    revoked_at       TIMESTAMPTZ,
    sdk_version      VARCHAR(20)  NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_consents_level CHECK (
        consent_level IN ('individual', 'aggregate', 'none')
    ),
    CONSTRAINT chk_consents_timeline CHECK (
        granted_at IS NULL OR revoked_at IS NULL OR granted_at <= revoked_at
    )
);

CREATE INDEX idx_consents_user_id ON consents (user_id);
CREATE INDEX idx_consents_level ON consents (consent_level);
CREATE INDEX idx_consents_active ON consents (is_active);

CREATE TABLE pseudonymous_mappings (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pseudonymous_id  UUID NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_pseudonymous_mappings_user UNIQUE (user_id),
    CONSTRAINT uq_pseudonymous_mappings_pseudo UNIQUE (pseudonymous_id)
);

CREATE INDEX idx_pseudonymous_mappings_user ON pseudonymous_mappings (user_id);
