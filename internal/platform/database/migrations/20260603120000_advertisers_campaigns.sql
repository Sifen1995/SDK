-- migrate:up

DROP TABLE IF EXISTS campaign_delivery_logs CASCADE;
DROP TABLE IF EXISTS campaigns CASCADE;
DROP TABLE IF EXISTS portal_users CASCADE;
DROP TABLE IF EXISTS advertisers CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS creatives CASCADE;

-- 1. ROLES
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (slug, display_name) VALUES
    ('operator_admin', 'Operator Admin'),
    ('advertiser', 'Advertiser'),
    ('read_only_analyst', 'Read-Only Analyst');

-- 2. ADVERTISERS (company account — no auth credentials; SDK keys live in developer portal)
CREATE TABLE advertisers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. PORTAL USERS (ad portal login; role via role_id)
CREATE TABLE portal_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    advertiser_id UUID REFERENCES advertisers(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_portal_users_email ON portal_users(email);
CREATE INDEX idx_portal_users_role ON portal_users(role_id);
CREATE INDEX idx_portal_users_advertiser ON portal_users(advertiser_id);

-- 4. CAMPAIGNS (creative payload embedded)
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_intent VARCHAR(100) NOT NULL,
    creative_format VARCHAR(50) NOT NULL,
    CONSTRAINT chk_campaign_creative_format CHECK (creative_format IN ('BANNER', 'PUSH_PLUS', 'SMS_PLUS')),
    title VARCHAR(255),
    body_text TEXT,
    image_url TEXT,
    destination_url TEXT NOT NULL,
    canvas_json JSONB NOT NULL DEFAULT '{}',
    daily_budget_cap NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    total_budget_cap NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT false,
    validation_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    validation_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_campaigns_intent_active ON campaigns(target_intent) WHERE is_active = true;
CREATE INDEX idx_campaigns_advertiser ON campaigns(advertiser_id);

-- 5. DELIVERY LOGS
CREATE TABLE campaign_delivery_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    delivery_status VARCHAR(50) NOT NULL,
    CONSTRAINT chk_delivery_status CHECK (delivery_status IN ('DISPATCHED', 'RENDERED', 'CLICKED', 'CONVERTED')),
    logged_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_delivery_campaign ON campaign_delivery_logs(campaign_id);

-- migrate:down

DROP TABLE IF EXISTS campaign_delivery_logs CASCADE;
DROP TABLE IF EXISTS campaigns CASCADE;
DROP TABLE IF EXISTS portal_users CASCADE;
DROP TABLE IF EXISTS advertisers CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
