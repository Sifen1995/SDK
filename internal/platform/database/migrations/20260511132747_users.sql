-- migrate:up

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    external_user_id  VARCHAR(255) NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_external_user_id ON users (external_user_id);

-- migrate:down
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "pgcrypto";

