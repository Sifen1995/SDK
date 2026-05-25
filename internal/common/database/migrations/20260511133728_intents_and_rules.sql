-- migrate:up


CREATE TABLE intents (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    intent_name  VARCHAR(100)   NOT NULL,
    confidence   NUMERIC(4, 3)  NOT NULL,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_confidence CHECK (confidence >= 0.0 AND confidence <= 1.0)
);

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
DROP TABLE IF EXISTS reward_rules;
DROP TABLE IF EXISTS intents;


