-- migrate:up


CREATE TABLE user_intent_profiles (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    pseudonymous_id  VARCHAR(64)  NOT NULL REFERENCES users(pseudonymous_id) ON DELETE CASCADE ON UPDATE CASCADE,
    intent_name      VARCHAR(100) NOT NULL,
    confidence       NUMERIC(4,3) NOT NULL CHECK (confidence BETWEEN 0.0 AND 1.0),
    model_version    VARCHAR(20)  NOT NULL,
    recorded_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ  NOT NULL,
    CONSTRAINT chk_profile_expiry CHECK (expires_at > recorded_at)
);

CREATE INDEX idx_user_intent_profiles_pseudo ON user_intent_profiles (pseudonymous_id);
CREATE INDEX idx_user_intent_profiles_recorded_at ON user_intent_profiles (recorded_at DESC);

CREATE TABLE intent_aggregate_counts (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_name    VARCHAR(100)  NOT NULL,
    date_bucket    DATE          NOT NULL DEFAULT CURRENT_DATE,
    signal_count   INTEGER       NOT NULL DEFAULT 0,
    weighted_count NUMERIC(10,2) NOT NULL DEFAULT 0.00,

    CONSTRAINT uq_intent_date UNIQUE (intent_name, date_bucket)
);

CREATE INDEX idx_agg_intent_date ON intent_aggregate_counts (intent_name, date_bucket DESC);

CREATE TABLE reward_rules (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_name  VARCHAR(100)   NOT NULL UNIQUE,
    reward_type  VARCHAR(50)    NOT NULL,
    amount       NUMERIC(10, 2) NOT NULL,
    currency     VARCHAR(50)    NOT NULL,
    message      TEXT           NOT NULL,
    is_active    BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- migrate:down
DROP TABLE IF EXISTS user_intent_profiles;
DROP TABLE IF EXISTS intent_aggregate_counts;
DROP TABLE IF EXISTS reward_rules;


