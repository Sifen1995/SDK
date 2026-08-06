package persistance

import (
	"time"

	geodomain "skykin-platform/internal/geofencing/domain"
)

type GeofenceZoneRow struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	AdvertiserID string    `gorm:"column:advertiser_id;type:uuid;not null;index"`
	Latitude     float64   `gorm:"column:latitude;type:numeric(10,7);not null"`
	Longitude    float64   `gorm:"column:longitude;type:numeric(10,7);not null"`
	RadiusMetres int       `gorm:"column:radius_metres;not null;default:100"`
	IsActive     bool      `gorm:"column:is_active;not null;default:false"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:now()"`
}

func (GeofenceZoneRow) TableName() string { return "geofence_zones" }

func (r GeofenceZoneRow) ToDomain() geodomain.GeofenceZone {
	return geodomain.GeofenceZone{
		ID:           r.ID,
		AdvertiserID: r.AdvertiserID,
		Latitude:     r.Latitude,
		Longitude:    r.Longitude,
		RadiusMetres: r.RadiusMetres,
		IsActive:     r.IsActive,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func GeofenceZoneRowFromDomain(z geodomain.GeofenceZone) GeofenceZoneRow {
	return GeofenceZoneRow{
		ID:           z.ID,
		AdvertiserID: z.AdvertiserID,
		Latitude:     z.Latitude,
		Longitude:    z.Longitude,
		RadiusMetres: z.RadiusMetres,
		IsActive:     z.IsActive,
		CreatedAt:    z.CreatedAt,
		UpdatedAt:    z.UpdatedAt,
	}
}

type CampaignGeofenceTargetRow struct {
	CampaignID string `gorm:"column:campaign_id;type:uuid;primaryKey"`
	ZoneID     string `gorm:"column:zone_id;type:uuid;primaryKey"`
}

func (CampaignGeofenceTargetRow) TableName() string { return "campaign_geofence_targets" }

type StoreVisitRow struct {
	ID             string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	PseudonymousID string    `gorm:"column:pseudonymous_id;type:uuid;not null;index"`
	ZoneID         string    `gorm:"column:zone_id;type:uuid;not null;index"`
	AccuracyM      *float64  `gorm:"column:accuracy_m;type:numeric(8,2)"`
	IsFlagged      bool      `gorm:"column:is_flagged;not null;default:false"`
	FlagReason     string    `gorm:"column:flag_reason;type:varchar(255)"`
	VisitedAt      time.Time `gorm:"column:visited_at;not null;default:now()"`
}

func (StoreVisitRow) TableName() string { return "store_visits" }

func (r StoreVisitRow) ToDomain() geodomain.StoreVisit {
	return geodomain.StoreVisit{
		ID:             r.ID,
		PseudonymousID: r.PseudonymousID,
		ZoneID:         r.ZoneID,
		AccuracyM:      r.AccuracyM,
		IsFlagged:      r.IsFlagged,
		FlagReason:     r.FlagReason,
		VisitedAt:      r.VisitedAt,
	}
}

func StoreVisitRowFromDomain(v geodomain.StoreVisit) StoreVisitRow {
	return StoreVisitRow{
		ID:             v.ID,
		PseudonymousID: v.PseudonymousID,
		ZoneID:         v.ZoneID,
		AccuracyM:      v.AccuracyM,
		IsFlagged:      v.IsFlagged,
		FlagReason:     v.FlagReason,
		VisitedAt:      v.VisitedAt,
	}
}
