-- migrate:up

-- Drop MVP tables replaced by this schema (safe if never created)
DROP TABLE IF EXISTS campaign_delivery_logs CASCADE;
DROP TABLE IF EXISTS creatives CASCADE;
DROP TABLE IF EXISTS campaigns CASCADE;
DROP TABLE IF EXISTS portal_users CASCADE;

-- 1. ADVERTISERS TABLE
CREATE TABLE advertisers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    api_key VARCHAR(64) NOT NULL UNIQUE,
    -- Ad portal RBAC (operator_admin | advertiser | read_only_analyst)
    role VARCHAR(50) NOT NULL DEFAULT 'advertiser',
    contact_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_advertisers_email ON advertisers(email);
CREATE INDEX idx_advertisers_role ON advertisers(role);

-- 2. CAMPAIGNS (creative payload embedded in campaign row)
CREATE TABLE campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,

    target_intent VARCHAR(100) NOT NULL,
    application_id VARCHAR(255) NOT NULL,

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

CREATE INDEX idx_campaigns_intent_app ON campaigns(target_intent, application_id) WHERE is_active = true;
CREATE INDEX idx_campaigns_advertiser ON campaigns(advertiser_id);

-- 3. CAMPAIGN METRICS / DELIVERY LOGS
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
DROP TABLE IF EXISTS advertisers CASCADE;
