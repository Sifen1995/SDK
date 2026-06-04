-- migrate:up
-- Fixes databases created by older GORM AutoMigrate (advertisers had email/password/api_key on one table).

-- Ensure roles exist
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (slug, display_name) VALUES
    ('operator_admin', 'Operator Admin'),
    ('advertiser', 'Advertiser'),
    ('read_only_analyst', 'Read-Only Analyst')
ON CONFLICT (slug) DO NOTHING;

-- Strip legacy auth columns from advertisers (company-only table)
ALTER TABLE advertisers DROP COLUMN IF EXISTS email;
ALTER TABLE advertisers DROP COLUMN IF EXISTS password_hash;
ALTER TABLE advertisers DROP COLUMN IF EXISTS api_key;
ALTER TABLE advertisers DROP COLUMN IF EXISTS role;
ALTER TABLE advertisers DROP COLUMN IF EXISTS contact_name;
ALTER TABLE advertisers DROP COLUMN IF EXISTS is_active;

ALTER TABLE advertisers ADD COLUMN IF NOT EXISTS company_name VARCHAR(255);
UPDATE advertisers SET company_name = 'Migrated Company' WHERE company_name IS NULL;
ALTER TABLE advertisers ALTER COLUMN company_name SET NOT NULL;

-- migrate:down
-- No down migration (schema alignment only)
