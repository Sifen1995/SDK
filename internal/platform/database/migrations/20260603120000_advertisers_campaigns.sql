-- migrate:up
-- Ad portal core tables (roles, advertisers, portal_users, campaigns, delivery).

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (slug, display_name) VALUES
    ('operator_admin', 'Operator Admin'),
    ('advertiser', 'Advertiser'),
    ('read_only_analyst', 'Read-Only Analyst')
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE IF NOT EXISTS advertisers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS portal_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    advertiser_id UUID REFERENCES advertisers(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_portal_users_email ON portal_users(email);
CREATE INDEX IF NOT EXISTS idx_portal_users_role ON portal_users(role_id);
CREATE INDEX IF NOT EXISTS idx_portal_users_advertiser ON portal_users(advertiser_id);

CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_intent VARCHAR(100) NOT NULL,
    channel_id UUID,
    title VARCHAR(255),
    body_text TEXT,
    image_url TEXT,
    canvas_json JSONB NOT NULL DEFAULT '{}',
    daily_budget_cap NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    total_budget_cap NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT false,
    validation_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    validation_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_campaigns_intent_active ON campaigns(target_intent) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_campaigns_advertiser ON campaigns(advertiser_id);

CREATE TABLE IF NOT EXISTS campaign_delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    delivery_status VARCHAR(50) NOT NULL,
    CONSTRAINT chk_delivery_status CHECK (delivery_status IN ('DISPATCHED', 'RENDERED', 'CLICKED', 'CONVERTED')),
    logged_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_delivery_campaign ON campaign_delivery_logs(campaign_id);

CREATE TABLE IF NOT EXISTS delivery_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_delivery_jobs_user_campaign ON delivery_jobs(user_id, campaign_id);
CREATE INDEX IF NOT EXISTS idx_delivery_jobs_user_campaign_day ON delivery_jobs(user_id, campaign_id, created_at);

-- migrate:down
DROP TABLE IF EXISTS delivery_jobs CASCADE;
DROP TABLE IF EXISTS campaign_delivery_logs CASCADE;
DROP TABLE IF EXISTS campaigns CASCADE;
DROP TABLE IF EXISTS portal_users CASCADE;
DROP TABLE IF EXISTS advertisers CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
