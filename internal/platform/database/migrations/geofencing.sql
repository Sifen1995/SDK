-- Geofencing schema (idempotent). Requires PostGIS.
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS geofence_zones (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id  UUID          NOT NULL,
    latitude       NUMERIC(10,7) NOT NULL,
    longitude      NUMERIC(10,7) NOT NULL,
    radius_metres  INTEGER       NOT NULL DEFAULT 100,
    is_active      BOOLEAN       NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_geofence_radius CHECK (radius_metres > 0 AND radius_metres <= 50000),
    CONSTRAINT chk_geofence_lat CHECK (latitude BETWEEN -90 AND 90),
    CONSTRAINT chk_geofence_lng CHECK (longitude BETWEEN -180 AND 180)
);

-- Geography point for ST_DWithin (lon, lat). Added separately so upgrades stay idempotent.
ALTER TABLE geofence_zones
    ADD COLUMN IF NOT EXISTS location geography(Point, 4326);

-- Backfill / refresh location from lat/lng for rows missing a point.
UPDATE geofence_zones
SET location = ST_SetSRID(
        ST_MakePoint(longitude::double precision, latitude::double precision),
        4326
    )::geography
WHERE location IS NULL;

CREATE OR REPLACE FUNCTION geofence_zones_set_location()
RETURNS TRIGGER AS $$
BEGIN
    NEW.location := ST_SetSRID(
        ST_MakePoint(NEW.longitude::double precision, NEW.latitude::double precision),
        4326
    )::geography;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_geofence_zones_location ON geofence_zones;
CREATE TRIGGER trg_geofence_zones_location
    BEFORE INSERT OR UPDATE OF latitude, longitude ON geofence_zones
    FOR EACH ROW EXECUTE FUNCTION geofence_zones_set_location();

CREATE INDEX IF NOT EXISTS idx_geofence_active
    ON geofence_zones (is_active)
    WHERE is_active = TRUE;

-- Existing installs may still default new rows to active; keep advertiser drafts inactive.
ALTER TABLE geofence_zones
    ALTER COLUMN is_active SET DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_geofence_advertiser
    ON geofence_zones (advertiser_id);

CREATE INDEX IF NOT EXISTS idx_geofence_location
    ON geofence_zones USING GIST (location);

-- Links campaigns to the zones they target (M:N).
CREATE TABLE IF NOT EXISTS campaign_geofence_targets (
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    zone_id     UUID NOT NULL REFERENCES geofence_zones(id) ON DELETE CASCADE,
    PRIMARY KEY (campaign_id, zone_id)
);

CREATE INDEX IF NOT EXISTS idx_cgt_zone ON campaign_geofence_targets (zone_id);

-- Records zone entry events from the SDK (demo + production visits).
CREATE TABLE IF NOT EXISTS store_visits (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    pseudonymous_id UUID          NOT NULL,
    zone_id         UUID          NOT NULL REFERENCES geofence_zones(id),
    accuracy_m      NUMERIC(8,2),
    is_flagged      BOOLEAN       NOT NULL DEFAULT FALSE,
    flag_reason     VARCHAR(255),
    visited_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_store_visits_pseudo
    ON store_visits (pseudonymous_id, visited_at DESC);
CREATE INDEX IF NOT EXISTS idx_store_visits_zone
    ON store_visits (zone_id);

-- Demo-only location ad consent (same cohort as SMS recipients).
ALTER TABLE demo_sms_recipients
    ADD COLUMN IF NOT EXISTS location_ad_consent BOOLEAN NOT NULL DEFAULT FALSE;
