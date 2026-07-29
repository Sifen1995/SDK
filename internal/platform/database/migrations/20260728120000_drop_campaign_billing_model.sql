-- Billing models belong to individual telemetry events, not campaigns.
ALTER TABLE campaigns DROP COLUMN IF EXISTS billing_model;
