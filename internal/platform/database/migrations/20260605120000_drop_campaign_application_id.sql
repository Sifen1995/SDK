-- migrate:up
ALTER TABLE campaigns DROP COLUMN IF EXISTS application_id;
DROP INDEX IF EXISTS idx_campaigns_intent_app;

CREATE INDEX IF NOT EXISTS idx_campaigns_intent_active
  ON campaigns(target_intent) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS delivery_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_delivery_jobs_user_campaign
  ON delivery_jobs(user_id, campaign_id);
CREATE INDEX IF NOT EXISTS idx_delivery_jobs_user_campaign_day
  ON delivery_jobs(user_id, campaign_id, created_at);

-- migrate:down
DROP TABLE IF EXISTS delivery_jobs;
DROP INDEX IF EXISTS idx_campaigns_intent_active;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS application_id VARCHAR(255);
