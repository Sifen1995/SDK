-- migrate:up

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    external_user_id  VARCHAR(255) NOT NULL UNIQUE,
    phone_number      VARCHAR(20)  NULL, -- 👈 Added optional (nullable) field
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_users_external_user_id ON users (external_user_id);
CREATE INDEX idx_users_phone_number ON users (phone_number); -- 👈 Added index for fast lookups

-- migrate:down
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "pgcrypto";