CREATE TABLE IF NOT EXISTS demo_sms_recipients (
    user_id              BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    phone_e164           TEXT        NOT NULL UNIQUE,
    display_name         TEXT,
    is_active            BOOLEAN     NOT NULL DEFAULT true,
    is_mock              BOOLEAN     NOT NULL DEFAULT true,
    provider_external_id TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sms_send_attempts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    send_key            TEXT        NOT NULL UNIQUE,
    campaign_id         UUID        NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    pseudonymous_id     UUID        NOT NULL,
    user_id             BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_e164          TEXT        NOT NULL,
    provider            VARCHAR(50) NOT NULL,
    provider_message_id TEXT,
    status              VARCHAR(50) NOT NULL,
    message_body        TEXT        NOT NULL,
    tracking_token      TEXT        NOT NULL UNIQUE,
    destination_url     TEXT        NOT NULL,
    error_message       TEXT,
    sent_at             TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_sms_send_attempt_status CHECK (status IN ('queued', 'sent', 'failed', 'delivered', 'clicked'))
);

CREATE INDEX IF NOT EXISTS idx_sms_send_attempts_campaign_id
    ON sms_send_attempts(campaign_id);
CREATE INDEX IF NOT EXISTS idx_sms_send_attempts_pseudonymous_id
    ON sms_send_attempts(pseudonymous_id);
CREATE INDEX IF NOT EXISTS idx_sms_send_attempts_user_id
    ON sms_send_attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_sms_send_attempts_status
    ON sms_send_attempts(status);
