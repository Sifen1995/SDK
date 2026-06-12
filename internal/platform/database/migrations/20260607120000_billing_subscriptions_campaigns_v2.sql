-- migrate:up
-- Billing, subscriptions, channels, audience segments, and campaign v2 columns.

-- ── CHANNELS ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS channels (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_premium  BOOLEAN      NOT NULL DEFAULT false,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO channels (code, name, description, is_premium) VALUES
    ('IN_APP_BANNER', 'In-App Banner',       'Banner ads shown inside host apps', false),
    ('PUSH',          'Push Notification',   'Push notification delivery',        false),
    ('SMS_PLUS',      'SMS+',                'Rich SMS with image and CTA',       true),
    ('NATIVE_FEED',   'Native Feed',         'Native in-feed ad units',           false)
ON CONFLICT (code) DO NOTHING;

-- ── SUBSCRIPTION PLANS ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS subscription_plans (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(100) NOT NULL UNIQUE,
    monthly_fee_etb      NUMERIC(12, 2) NOT NULL,
    max_active_campaigns INT            NOT NULL DEFAULT 3,
    max_daily_budget_etb NUMERIC(12, 2) NOT NULL,
    included_impressions INT            NOT NULL DEFAULT 0,
    sms_plus_enabled     BOOLEAN        NOT NULL DEFAULT false,
    audiencemart_enabled BOOLEAN        NOT NULL DEFAULT false,
    cpc_discount_pct     NUMERIC(5, 2)  DEFAULT 0,
    is_active            BOOLEAN        NOT NULL DEFAULT true,
    created_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

INSERT INTO subscription_plans (
    name, monthly_fee_etb, max_active_campaigns, max_daily_budget_etb,
    included_impressions, sms_plus_enabled, audiencemart_enabled, cpc_discount_pct
) VALUES
    ('Starter',    5000.00,  3,   500.00,  10000,  false, false,  0.00),
    ('Growth',    15000.00, 10,  2000.00,  50000,  true,  true,   5.00),
    ('Enterprise',50000.00, 100, 10000.00, 200000, true,  true,  15.00)
ON CONFLICT (name) DO NOTHING;

-- ── ADVERTISER SUBSCRIPTIONS ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS advertiser_subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id        UUID         NOT NULL UNIQUE REFERENCES advertisers(id) ON DELETE CASCADE,
    plan_id              UUID         NOT NULL REFERENCES subscription_plans(id),
    status               VARCHAR(20)  NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ  NOT NULL,
    current_period_end   TIMESTAMPTZ  NOT NULL,
    impressions_used     INT          NOT NULL DEFAULT 0,
    cancelled_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_advertiser_subscription_status
        CHECK (status IN ('active', 'suspended', 'cancelled', 'past_due'))
);

CREATE INDEX IF NOT EXISTS idx_advertiser_subscriptions_plan
    ON advertiser_subscriptions(plan_id);

-- ── BILLING RATES ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS billing_rates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id    UUID          NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    event_type VARCHAR(50)   NOT NULL,
    model      VARCHAR(10)   NOT NULL,
    rate_etb   NUMERIC(10, 4) NOT NULL,
    is_active  BOOLEAN       NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_billing_rate_event_type
        CHECK (event_type IN ('impression', 'click', 'install', 'signup', 'purchase')),
    CONSTRAINT chk_billing_rate_model
        CHECK (model IN ('CPM', 'CPC', 'CPI', 'CPA', 'REV_SHARE')),
    CONSTRAINT uq_billing_rates_plan_event_model
        UNIQUE (plan_id, event_type, model)
);

CREATE INDEX IF NOT EXISTS idx_billing_rates_plan ON billing_rates(plan_id);

-- Seed default CPC/CPM rates per plan
INSERT INTO billing_rates (plan_id, event_type, model, rate_etb)
SELECT p.id, v.event_type, v.model, v.rate_etb
FROM subscription_plans p
CROSS JOIN (VALUES
    ('impression', 'CPM',  2.5000),
    ('click',      'CPC',  0.7500),
    ('install',    'CPI', 15.0000),
    ('signup',     'CPA', 25.0000),
    ('purchase',   'REV_SHARE', 5.0000)
) AS v(event_type, model, rate_etb)
ON CONFLICT (plan_id, event_type, model) DO NOTHING;

-- ── BILLING EVENTS ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS billing_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id     UUID           NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    campaign_id       UUID           NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    subscription_id   UUID           NOT NULL REFERENCES advertiser_subscriptions(id) ON DELETE CASCADE,
    event_type        VARCHAR(50)    NOT NULL,
    billing_model     VARCHAR(20)    NOT NULL,
    rate_applied      NUMERIC(10, 4) NOT NULL,
    transaction_value NUMERIC(12, 2) DEFAULT 0,
    charge_etb        NUMERIC(10, 4) NOT NULL,
    is_billed         BOOLEAN        NOT NULL DEFAULT false,
    occurred_at       TIMESTAMPTZ    NOT NULL,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_billing_event_type
        CHECK (event_type IN ('impression', 'click', 'install', 'signup', 'purchase')),
    CONSTRAINT chk_billing_event_model
        CHECK (billing_model IN ('CPM', 'CPC', 'CPI', 'CPA', 'REV_SHARE'))
);

CREATE INDEX IF NOT EXISTS idx_billing_events_advertiser ON billing_events(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_campaign  ON billing_events(campaign_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_unbilled  ON billing_events(is_billed) WHERE is_billed = false;
CREATE INDEX IF NOT EXISTS idx_billing_events_occurred  ON billing_events(occurred_at);

-- ── INVOICES ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS invoices (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id        UUID          NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    subscription_id      UUID          NOT NULL REFERENCES advertiser_subscriptions(id) ON DELETE CASCADE,
    period_start         TIMESTAMPTZ   NOT NULL,
    period_end           TIMESTAMPTZ   NOT NULL,
    subscription_fee_etb NUMERIC(12, 2) NOT NULL,
    usage_fee_etb        NUMERIC(12, 2) NOT NULL,
    total_etb            NUMERIC(12, 2) NOT NULL,
    status               VARCHAR(20)   NOT NULL DEFAULT 'draft',
    paid_at              TIMESTAMPTZ,
    payment_ref          VARCHAR(255),
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_invoice_status
        CHECK (status IN ('draft', 'sent', 'paid', 'overdue', 'void'))
);

CREATE INDEX IF NOT EXISTS idx_invoices_advertiser ON invoices(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_invoices_period     ON invoices(period_start, period_end);

-- ── AUDIENCE SEGMENTS (Audiencemart) ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS audience_segments (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name               VARCHAR(255) NOT NULL,
    description        TEXT,
    top_intent_signals JSONB        NOT NULL DEFAULT '[]',
    approximate_size   INT          NOT NULL DEFAULT 0,
    estimated_cpm      NUMERIC(10, 2) NOT NULL,
    available_from     TIMESTAMPTZ  NOT NULL,
    available_until    TIMESTAMPTZ,
    is_active          BOOLEAN      NOT NULL DEFAULT true,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── SEGMENT PURCHASES ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS segment_purchases (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id UUID          NOT NULL REFERENCES advertisers(id) ON DELETE CASCADE,
    segment_id    UUID          NOT NULL REFERENCES audience_segments(id) ON DELETE CASCADE,
    campaign_id   UUID          NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    amount_paid   NUMERIC(12, 2) NOT NULL,
    valid_from    TIMESTAMPTZ   NOT NULL,
    valid_until   TIMESTAMPTZ   NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_segment_purchases_advertiser ON segment_purchases(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_segment_purchases_segment    ON segment_purchases(segment_id);
CREATE INDEX IF NOT EXISTS idx_segment_purchases_campaign   ON segment_purchases(campaign_id);

-- ── CAMPAIGNS v2 (additive — creative_format kept until app code migrates) ──
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS channel_id UUID REFERENCES channels(id);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS segment_id UUID REFERENCES audience_segments(id);
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS destination_url TEXT;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS billing_model VARCHAR(20) NOT NULL DEFAULT 'CPC';
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS budget_spent NUMERIC(12, 2) NOT NULL DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS frequency_cap_per_day INT NOT NULL DEFAULT 3;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS scheduled_start_at TIMESTAMPTZ;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS scheduled_end_at TIMESTAMPTZ;

ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS chk_campaign_billing_model;
ALTER TABLE campaigns ADD CONSTRAINT chk_campaign_billing_model
    CHECK (billing_model IN ('CPM', 'CPC', 'CPI', 'CPA', 'REV_SHARE'));

-- Backfill channel_id from legacy creative_format
UPDATE campaigns c
SET channel_id = ch.id
FROM channels ch
WHERE c.channel_id IS NULL
  AND (
    (c.creative_format = 'BANNER'     AND ch.code = 'IN_APP_BANNER') OR
    (c.creative_format = 'PUSH_PLUS'  AND ch.code = 'PUSH') OR
    (c.creative_format = 'SMS_PLUS'     AND ch.code = 'SMS_PLUS')
  );

UPDATE campaigns
SET channel_id = (SELECT id FROM channels WHERE code = 'IN_APP_BANNER' LIMIT 1)
WHERE channel_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_campaigns_channel_id ON campaigns(channel_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_segment_id ON campaigns(segment_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_start ON campaigns(scheduled_start_at);
CREATE INDEX IF NOT EXISTS idx_campaigns_scheduled_end   ON campaigns(scheduled_end_at);

-- migrate:down
ALTER TABLE campaigns DROP COLUMN IF EXISTS scheduled_end_at;
ALTER TABLE campaigns DROP COLUMN IF EXISTS scheduled_start_at;
ALTER TABLE campaigns DROP COLUMN IF EXISTS frequency_cap_per_day;
ALTER TABLE campaigns DROP COLUMN IF EXISTS budget_spent;
ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS chk_campaign_billing_model;
ALTER TABLE campaigns DROP COLUMN IF EXISTS billing_model;
ALTER TABLE campaigns DROP COLUMN IF EXISTS destination_url;
ALTER TABLE campaigns DROP COLUMN IF EXISTS segment_id;
ALTER TABLE campaigns DROP COLUMN IF EXISTS channel_id;

DROP TABLE IF EXISTS segment_purchases CASCADE;
DROP TABLE IF EXISTS audience_segments CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS billing_events CASCADE;
DROP TABLE IF EXISTS billing_rates CASCADE;
DROP TABLE IF EXISTS advertiser_subscriptions CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;
DROP TABLE IF EXISTS channels CASCADE;
