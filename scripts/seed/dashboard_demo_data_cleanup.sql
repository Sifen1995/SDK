-- Dashboard Demo Data Cleanup Script
-- Safely removes development/demo data for the Skykin Portals by deleting records
-- associated with the hardcoded Advertiser UUID defined in the seed script.

DO $$
DECLARE
    v_advertiser_id UUID := '11111111-1111-4111-8111-111111111111';
BEGIN
    -- 1. Delete analytics data (Delivery Jobs and Logs)
    DELETE FROM campaign_delivery_logs WHERE campaign_id IN (SELECT id FROM campaigns WHERE advertiser_id = v_advertiser_id);
    DELETE FROM delivery_jobs WHERE campaign_id IN (SELECT id FROM campaigns WHERE advertiser_id = v_advertiser_id);
    
    -- 2. Delete campaigns
    DELETE FROM campaigns WHERE advertiser_id = v_advertiser_id;
    
    -- 3. Delete subscriptions
    DELETE FROM advertiser_subscriptions WHERE advertiser_id = v_advertiser_id;
    
    -- 4. Delete advertiser
    DELETE FROM advertisers WHERE id = v_advertiser_id;
END $$;
