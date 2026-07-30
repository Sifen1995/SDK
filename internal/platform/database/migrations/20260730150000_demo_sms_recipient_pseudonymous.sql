ALTER TABLE demo_sms_recipients
    ADD COLUMN IF NOT EXISTS pseudonymous_id UUID;

UPDATE demo_sms_recipients rec
SET pseudonymous_id = pm.pseudonymous_id
FROM pseudonymous_mappings pm
WHERE pm.user_id = rec.user_id
  AND (rec.pseudonymous_id IS NULL OR rec.pseudonymous_id <> pm.pseudonymous_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_demo_sms_recipients_pseudonymous_id
    ON demo_sms_recipients (pseudonymous_id)
    WHERE pseudonymous_id IS NOT NULL;
