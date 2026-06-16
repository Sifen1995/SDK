-- migrate:up
-- Operator moderation queue + remove legacy creative_format (channel_id is canonical).

ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(20) NOT NULL DEFAULT 'pending';
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderation_notes TEXT;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMPTZ;
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderated_by UUID;

ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS chk_campaign_moderation_status;
ALTER TABLE campaigns ADD CONSTRAINT chk_campaign_moderation_status
    CHECK (moderation_status IN ('pending', 'approved', 'rejected'));

-- Existing live campaigns are treated as already approved.
UPDATE campaigns SET moderation_status = 'approved' WHERE is_active = true;
UPDATE campaigns SET moderation_status = 'pending' WHERE moderation_status IS NULL OR moderation_status = 'pending';

ALTER TABLE campaigns DROP COLUMN IF EXISTS creative_format;
ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS chk_campaign_creative_format;

CREATE INDEX IF NOT EXISTS idx_campaigns_moderation_pending
    ON campaigns(moderation_status) WHERE moderation_status = 'pending' AND is_active = false;

-- migrate:down
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS creative_format VARCHAR(50);
ALTER TABLE campaigns DROP COLUMN IF EXISTS moderated_by;
ALTER TABLE campaigns DROP COLUMN IF EXISTS moderated_at;
ALTER TABLE campaigns DROP COLUMN IF EXISTS moderation_notes;
ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS chk_campaign_moderation_status;
ALTER TABLE campaigns DROP COLUMN IF EXISTS moderation_status;
DROP INDEX IF EXISTS idx_campaigns_moderation_pending;
