-- Curated fraud intelligence distributed to mobile clients.
-- This migration is deliberately idempotent because Migrate executes it at startup.
CREATE TABLE IF NOT EXISTS blocked_domains (
    domain          VARCHAR(255) PRIMARY KEY,
    threat_type     VARCHAR(64) NOT NULL, -- url_phishing | financial_scam | brand_impersonation
    severity        VARCHAR(32) NOT NULL, -- low | medium | high | critical
    source          VARCHAR(64) NOT NULL, -- community_report | manual_review | auto_detected
    status          VARCHAR(32) NOT NULL DEFAULT 'active', -- active | revoked
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blocked_senders (
    sender_hash     VARCHAR(64) PRIMARY KEY, -- SHA256(phone_number) to prevent storing raw numbers
    threat_type     VARCHAR(64) NOT NULL,
    severity        VARCHAR(32) NOT NULL,
    source          VARCHAR(64) NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'active', -- active | revoked
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Upgrade installations created by the earlier schema.
ALTER TABLE blocked_domains
    ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

ALTER TABLE blocked_senders
    ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE TABLE IF NOT EXISTS scam_patterns (
    id              VARCHAR(64) PRIMARY KEY,
    pattern_type    VARCHAR(32) NOT NULL, -- regex | keyword_combo | url_pattern
    pattern_value   TEXT NOT NULL,
    threat_category VARCHAR(64) NOT NULL,
    confidence      NUMERIC(3,2) NOT NULL,
    language        VARCHAR(8) NOT NULL DEFAULT 'any', -- en | am | any
    is_active       BOOLEAN NOT NULL DEFAULT true,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS threat_reports (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    threat_type      VARCHAR(64) NOT NULL,
    severity         VARCHAR(32) NOT NULL,
    sender_hash      VARCHAR(64) NULL,
    url_domain       VARCHAR(255) NULL,
    detection_source VARCHAR(32) NOT NULL, -- blocklist | pattern | ml
    sdk_version      VARCHAR(32) NOT NULL,
    reported_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add named checks only when absent.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_domain_severity') THEN
        ALTER TABLE blocked_domains
            ADD CONSTRAINT chk_domain_severity
            CHECK (severity IN ('low', 'medium', 'high', 'critical'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_domain_status') THEN
        ALTER TABLE blocked_domains
            ADD CONSTRAINT chk_domain_status CHECK (status IN ('active', 'revoked'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_sender_severity') THEN
        ALTER TABLE blocked_senders
            ADD CONSTRAINT chk_sender_severity
            CHECK (severity IN ('low', 'medium', 'high', 'critical'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_sender_status') THEN
        ALTER TABLE blocked_senders
            ADD CONSTRAINT chk_sender_status CHECK (status IN ('active', 'revoked'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_threat_report_type') THEN
        ALTER TABLE threat_reports
            ADD CONSTRAINT chk_threat_report_type
            CHECK (threat_type IN ('url_phishing', 'financial_scam', 'brand_impersonation'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_threat_report_severity') THEN
        ALTER TABLE threat_reports
            ADD CONSTRAINT chk_threat_report_severity
            CHECK (severity IN ('low', 'medium', 'high', 'critical'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_threat_report_source') THEN
        ALTER TABLE threat_reports
            ADD CONSTRAINT chk_threat_report_source
            CHECK (detection_source IN ('blocklist', 'pattern', 'ml'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_threat_report_indicator') THEN
        ALTER TABLE threat_reports
            ADD CONSTRAINT chk_threat_report_indicator
            CHECK (sender_hash IS NOT NULL OR url_domain IS NOT NULL);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_threat_report_sender_hash') THEN
        ALTER TABLE threat_reports
            ADD CONSTRAINT chk_threat_report_sender_hash
            CHECK (sender_hash IS NULL OR sender_hash ~ '^[0-9a-f]{64}$');
    END IF;
END $$;

-- Database triggers keep delta cursors correct even for raw SQL/admin updates.
CREATE OR REPLACE FUNCTION fraud_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_blocked_domains_updated_at ON blocked_domains;
CREATE TRIGGER trg_blocked_domains_updated_at
    BEFORE UPDATE ON blocked_domains
    FOR EACH ROW EXECUTE FUNCTION fraud_set_updated_at();

DROP TRIGGER IF EXISTS trg_blocked_senders_updated_at ON blocked_senders;
CREATE TRIGGER trg_blocked_senders_updated_at
    BEFORE UPDATE ON blocked_senders
    FOR EACH ROW EXECUTE FUNCTION fraud_set_updated_at();

DROP TRIGGER IF EXISTS trg_scam_patterns_updated_at ON scam_patterns;
CREATE TRIGGER trg_scam_patterns_updated_at
    BEFORE UPDATE ON scam_patterns
    FOR EACH ROW EXECUTE FUNCTION fraud_set_updated_at();

CREATE INDEX IF NOT EXISTS idx_blocked_domains_expires
    ON blocked_domains(expires_at)
    WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_blocked_domains_sync
    ON blocked_domains(updated_at, domain);
CREATE INDEX IF NOT EXISTS idx_blocked_senders_sync
    ON blocked_senders(updated_at, sender_hash);
CREATE INDEX IF NOT EXISTS idx_scam_patterns_sync
    ON scam_patterns(updated_at, id);
CREATE INDEX IF NOT EXISTS idx_threat_reports_domain
    ON threat_reports(url_domain, reported_at) WHERE url_domain IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_threat_reports_sender
    ON threat_reports(sender_hash, reported_at) WHERE sender_hash IS NOT NULL;