-- migrate:up


CREATE TABLE rewards (
    id           UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID           NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    intent_id    UUID           NOT NULL REFERENCES intents(id) ON DELETE CASCADE,
    rule_id      UUID           NOT NULL REFERENCES reward_rules(id),
    reward_type  VARCHAR(50)    NOT NULL,
    amount       NUMERIC(10, 2) NOT NULL,
    currency     VARCHAR(50)    NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'pending',
    message      TEXT           NOT NULL,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    sent_at      TIMESTAMPTZ,
    claimed_at   TIMESTAMPTZ,
    CONSTRAINT chk_reward_status CHECK (status IN ('pending', 'sent', 'claimed', 'expired'))
);

-- migrate:down
DROP TABLE IF EXISTS rewards;


