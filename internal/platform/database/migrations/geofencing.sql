-- internal/platform/database/migrations/geofencing.sql

CREATE TABLE geofence_zones (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id    VARCHAR(255)  NOT NULL,
    merchant_name  VARCHAR(255)  NOT NULL,
    latitude       NUMERIC(10,7) NOT NULL,
    longitude      NUMERIC(10,7) NOT NULL,
    radius_metres  INTEGER       NOT NULL DEFAULT 100,
    is_active      BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_geofence_active ON geofence_zones (is_active)
    WHERE is_active = TRUE;

-- Links campaigns to the zones they target
-- One campaign can target multiple zones
-- One zone can be targeted by multiple campaigns
CREATE TABLE campaign_geofence_targets (
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    zone_id     UUID NOT NULL REFERENCES geofence_zones(id) ON DELETE CASCADE,
    PRIMARY KEY (campaign_id, zone_id)
);

CREATE INDEX idx_cgt_zone ON campaign_geofence_targets (zone_id);

-- Records zone entry events from SDK
CREATE TABLE store_visits (
    id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL REFERENCES users(id),
    zone_id     UUID          NOT NULL REFERENCES geofence_zones(id),
    accuracy_m  NUMERIC(8,2),
    is_flagged  BOOLEAN       NOT NULL DEFAULT FALSE,
    flag_reason VARCHAR(255),
    visited_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_store_visits_user ON store_visits (user_id, visited_at DESC);
CREATE INDEX idx_store_visits_zone ON store_visits (zone_id);