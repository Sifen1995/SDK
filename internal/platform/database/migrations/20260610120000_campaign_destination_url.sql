-- Ensure destination_url exists on campaigns (click-through URL for creatives).
ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS destination_url TEXT;
