-- Align every end-user identity column on the pseudonymous id.
--
-- users.id (BIGINT) is the internal primary key and never leaves the platform.
-- pseudonymous_mappings translates it to the UUID the SDK sends. Only users,
-- consents and pseudonymous_mappings keep a user_id column; every table that
-- stores behavioural data now names the column pseudonymous_id so the type and
-- the meaning of the value agree.

DO $$
DECLARE
    tbl text;
BEGIN
    FOREACH tbl IN ARRAY ARRAY[
        'intents',
        'segment_memberships',
        'delivery_jobs',
        'campaign_delivery_logs',
        'rewards',
        'events',
        'store_visits'
    ]
    LOOP
        IF EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = tbl AND column_name = 'user_id'
        ) AND NOT EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_schema = 'public' AND table_name = tbl AND column_name = 'pseudonymous_id'
        ) THEN
            EXECUTE format('ALTER TABLE public.%I RENAME COLUMN user_id TO pseudonymous_id', tbl);
        END IF;
    END LOOP;
END $$;

-- segment_memberships declared a UUID column with a foreign key to users(id),
-- which is BIGINT. The constraint could never be satisfied, so membership rows
-- either failed to insert or the table was created without the constraint.
ALTER TABLE IF EXISTS segment_memberships
    DROP CONSTRAINT IF EXISTS segment_memberships_user_id_fkey;
ALTER TABLE IF EXISTS segment_memberships
    DROP CONSTRAINT IF EXISTS segment_memberships_pseudonymous_id_fkey;

-- intents held pseudonymous UUIDs in a varchar column that had been widened to
-- accept bigint ids as decimal strings. Drop those legacy rows and enforce UUID.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'intents'
          AND column_name = 'pseudonymous_id' AND data_type <> 'uuid'
    ) THEN
        DELETE FROM intents
        WHERE pseudonymous_id !~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$';
        ALTER TABLE intents
            ALTER COLUMN pseudonymous_id TYPE UUID USING pseudonymous_id::uuid;
    END IF;
END $$;

ALTER INDEX IF EXISTS idx_intents_user_id             RENAME TO idx_intents_pseudonymous_id;
ALTER INDEX IF EXISTS idx_rewards_user_id             RENAME TO idx_rewards_pseudonymous_id;
ALTER INDEX IF EXISTS idx_events_user_id              RENAME TO idx_events_pseudonymous_id;
ALTER INDEX IF EXISTS idx_delivery_jobs_user_id       RENAME TO idx_delivery_jobs_pseudonymous_id;
ALTER INDEX IF EXISTS idx_delivery_jobs_user_campaign RENAME TO idx_delivery_jobs_pseudonymous_campaign;
ALTER INDEX IF EXISTS idx_seg_memberships_user        RENAME TO idx_seg_memberships_pseudonymous_id;
ALTER INDEX IF EXISTS idx_store_visits_user           RENAME TO idx_store_visits_pseudonymous_id;
