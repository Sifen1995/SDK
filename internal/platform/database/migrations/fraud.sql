-- Track all curated domains distributed to mobile clients
CREATE TABLE blocked_domains (
    domain          VARCHAR(255) PRIMARY KEY,
    threat_type     VARCHAR(64) NOT NULL, -- url_phishing | financial_scam | brand_impersonation
    severity        VARCHAR(32) NOT NULL, -- low | medium | high | critical
    source          VARCHAR(64) NOT NULL, -- community_report | manual_review | auto_detected
    status          VARCHAR(32) NOT NULL DEFAULT 'active', -- active | revoked
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NULL
);

-- Track senders/shortcodes/phone numbers
CREATE TABLE blocked_senders (
    sender_hash     VARCHAR(64) PRIMARY KEY, -- SHA256(phone_number) to prevent storing raw numbers
    threat_type     VARCHAR(64) NOT NULL,
    severity        VARCHAR(32) NOT NULL,
    source          VARCHAR(64) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE blocked_domains 
    ADD CONSTRAINT chk_domain_severity CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    ADD CONSTRAINT chk_domain_status CHECK (status IN ('active', 'revoked'));

CREATE INDEX idx_blocked_domains_expires 
    ON blocked_domains(expires_at) 
    WHERE expires_at IS NOT NULL;

-- Scam detection patterns delivered to clients
CREATE TABLE scam_patterns (
    id              VARCHAR(64) PRIMARY KEY,
    pattern_type    VARCHAR(32) NOT NULL, -- regex | keyword_combo | url_pattern
    pattern_value   TEXT NOT NULL,
    threat_category VARCHAR(64) NOT NULL,
    confidence      NUMERIC(3,2) NOT NULL,
    language        VARCHAR(8) NOT NULL DEFAULT 'any', -- en | am | any
    is_active       BOOLEAN NOT NULL DEFAULT true,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Anonymous client reports submitted from devices
CREATE TABLE threat_reports (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   threat_type      VARCHAR(64) NOT NULL,
    severity         VARCHAR(32) NOT NULL,
    sender_hash      VARCHAR(64) NULL,
    url_domain       VARCHAR(255) NULL,
    detection_source VARCHAR(32) NOT NULL, -- blocklist | pattern | ml
    sdk_version      VARCHAR(32) NOT NULL,
    reported_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexing for sliding-window aggregation performance
CREATE INDEX idx_threat_reports_domain ON threat_reports(url_domain, reported_at) WHERE url_domain IS NOT NULL;
CREATE INDEX idx_threat_reports_sender ON threat_reports(sender_hash, reported_at) WHERE sender_hash IS NOT NULL;