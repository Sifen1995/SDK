-- Dashboard Demo Data Seed Script
-- Safely creates development/demo data for the Skykin Portals (Admin & Ad Portal).
-- Uses hardcoded UUIDs for the advertiser to ensure idempotent execution
-- and safe, reliable cleanup without affecting production data.

DO $$
DECLARE
    v_advertiser_id UUID := '11111111-1111-4111-8111-111111111111';
    v_plan_id UUID;
    v_channel_id UUID;
    v_campaign_id UUID;
    i INT;
    j INT;
    v_status TEXT;
    v_intent TEXT;
    v_user UUID;
BEGIN
    -- 1. Create Demo Advertiser
    IF NOT EXISTS (SELECT 1 FROM advertisers WHERE id = v_advertiser_id) THEN
        INSERT INTO advertisers (id, company_name, created_at, updated_at)
        VALUES (v_advertiser_id, 'Demo Advertiser (Seed)', NOW(), NOW());
    END IF;

    -- 2. Create Active Subscription
    SELECT id INTO v_plan_id FROM subscription_plans WHERE name = 'Growth' LIMIT 1;
    IF v_plan_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM advertiser_subscriptions WHERE advertiser_id = v_advertiser_id) THEN
        INSERT INTO advertiser_subscriptions (id, advertiser_id, plan_id, status, current_period_start, current_period_end, impressions_used, created_at, updated_at)
        VALUES (gen_random_uuid(), v_advertiser_id, v_plan_id, 'active', NOW() - INTERVAL '10 days', NOW() + INTERVAL '20 days', 1500, NOW(), NOW());
    END IF;

    -- 3. Create Many Campaigns
    SELECT id INTO v_channel_id FROM channels WHERE code = 'IN_APP_BANNER' LIMIT 1;
    
    IF v_channel_id IS NOT NULL THEN
        -- Clear old seed deliveries first for idempotency
        DELETE FROM campaign_delivery_logs WHERE campaign_id IN (SELECT id FROM campaigns WHERE advertiser_id = v_advertiser_id);
        DELETE FROM delivery_jobs WHERE campaign_id IN (SELECT id FROM campaigns WHERE advertiser_id = v_advertiser_id);
        DELETE FROM campaigns WHERE advertiser_id = v_advertiser_id;
        
        -- Generate 15 campaigns
        FOR i IN 1..15 LOOP
            v_campaign_id := gen_random_uuid();
            
            -- Randomize status
            IF i % 3 = 0 THEN
                v_status := 'pending';
            ELSIF i % 5 = 0 THEN
                v_status := 'rejected';
            ELSE
                v_status := 'approved';
            END IF;

            -- Randomize intent
            IF i % 2 = 0 THEN
                v_intent := 'gaming_interest';
            ELSE
                v_intent := 'fashion_interest';
            END IF;

            INSERT INTO campaigns (id, advertiser_id, name, target_intent, channel_id, title, body_text, billing_model, daily_budget_cap, total_budget_cap, budget_spent, is_active, moderation_status, validation_status, created_at, updated_at)
            VALUES (
                v_campaign_id, 
                v_advertiser_id, 
                'Demo Campaign ' || i, 
                v_intent, 
                v_channel_id, 
                'Summer Sale ' || i, 
                'Get 50% off all summer apparel.', 
                'CPC', 
                500, 
                5000, 
                i * 10, 
                v_status = 'approved', 
                v_status, 
                'valid', 
                NOW() - (i || ' days')::INTERVAL, 
                NOW()
            );

            -- 4. Create Delivery Jobs and Logs for approved campaigns
            IF v_status = 'approved' THEN
                -- Generate 10-20 delivery jobs per active campaign
                FOR j IN 1..(10 + (i % 10)) LOOP
                    v_user := gen_random_uuid();
                    
                    INSERT INTO delivery_jobs (id, user_id, campaign_id, created_at) 
                    VALUES (gen_random_uuid(), v_user, v_campaign_id, NOW() - (j || ' hours')::INTERVAL);
                    
                    INSERT INTO campaign_delivery_logs (id, campaign_id, user_id, session_id, delivery_status, logged_at) 
                    VALUES (gen_random_uuid(), v_campaign_id, v_user, 'sess_' || j, 'delivered', NOW() - (j || ' hours')::INTERVAL);
                    
                    -- Simulate clicks for some deliveries
                    IF j % 4 = 0 THEN
                        INSERT INTO campaign_delivery_logs (id, campaign_id, user_id, session_id, delivery_status, logged_at) 
                        VALUES (gen_random_uuid(), v_campaign_id, v_user, 'sess_' || j, 'clicked', NOW() - (j || ' hours')::INTERVAL + INTERVAL '1 minute');
                    END IF;
                END LOOP;
            END IF;
        END LOOP;
    END IF;
END $$;
