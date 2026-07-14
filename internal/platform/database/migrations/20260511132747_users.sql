-- migrate:up

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    pseudonymous_id  VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_pseudonymous_id ON users (pseudonymous_id);

-- migrate:down
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "pgcrypto";