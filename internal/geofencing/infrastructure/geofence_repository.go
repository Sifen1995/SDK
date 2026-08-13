package infrastructure

import (
	"context"
	"fmt"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignpersistence "skykin-platform/internal/campaigns/infrastructure/persistence"
	geodomain "skykin-platform/internal/geofencing/domain"
	"skykin-platform/internal/geofencing/infrastructure/persistance"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GeofenceRepository struct {
	db *gorm.DB
}

func NewGeofenceRepository(db *gorm.DB) *GeofenceRepository {
	return &GeofenceRepository{db: db}
}

var (
	_ geodomain.ZoneRepository            = (*GeofenceRepository)(nil)
	_ geodomain.TargetRepository          = (*GeofenceRepository)(nil)
	_ geodomain.LocationConsentRepository = (*GeofenceRepository)(nil)
)

func (r *GeofenceRepository) Create(ctx context.Context, zone *geodomain.GeofenceZone) error {
	if zone.ID == "" {
		zone.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if zone.CreatedAt.IsZero() {
		zone.CreatedAt = now
	}
	zone.UpdatedAt = now
	row := persistance.GeofenceZoneRowFromDomain(*zone)
	if err := r.db.WithContext(ctx).
		Select("ID", "AdvertiserID", "Latitude", "Longitude", "RadiusMetres", "IsActive", "CreatedAt", "UpdatedAt").
		Create(&row).Error; err != nil {
		return err
	}
	*zone = row.ToDomain()
	return nil
}

func (r *GeofenceRepository) ListByAdvertiser(ctx context.Context, advertiserID string) ([]geodomain.GeofenceZone, error) {
	var rows []persistance.GeofenceZoneRow
	if err := r.db.WithContext(ctx).
		Where("advertiser_id = ?", advertiserID).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]geodomain.GeofenceZone, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out, nil
}

func (r *GeofenceRepository) ListInactive(ctx context.Context) ([]geodomain.GeofenceZone, error) {
	var rows []persistance.GeofenceZoneRow
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", false).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]geodomain.GeofenceZone, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out, nil
}

func (r *GeofenceRepository) GetByID(ctx context.Context, zoneID string) (*geodomain.GeofenceZone, error) {
	var row persistance.GeofenceZoneRow
	if err := r.db.WithContext(ctx).Where("id = ?", zoneID).First(&row).Error; err != nil {
		return nil, err
	}
	zone := row.ToDomain()
	return &zone, nil
}

// ActivateZone activates a draft store. Already-active zones are returned unchanged.
func (r *GeofenceRepository) ActivateZone(ctx context.Context, zoneID string) (*geodomain.GeofenceZone, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE geofence_zones
		SET is_active = TRUE, updated_at = NOW()
		WHERE id = ? AND is_active = FALSE
	`, zoneID)
	if res.Error != nil {
		return nil, res.Error
	}
	return r.GetByID(ctx, zoneID)
}

// ActivateForCampaign activates inactive zones linked to the campaign (admin approval).
func (r *GeofenceRepository) ActivateForCampaign(ctx context.Context, campaignID string) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE geofence_zones AS z
		SET is_active = TRUE, updated_at = NOW()
		FROM campaign_geofence_targets AS t
		WHERE t.zone_id = z.id
		  AND t.campaign_id = ?
		  AND z.is_active = FALSE
	`, campaignID).Error
}

func (r *GeofenceRepository) FindNearby(
	ctx context.Context,
	lat, lng float64,
	radiusMetres int,
) ([]geodomain.GeofenceZone, error) {
	var rows []persistance.GeofenceZoneRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, advertiser_id, latitude, longitude, radius_metres, is_active, created_at, updated_at
		FROM geofence_zones
		WHERE is_active = TRUE
		  AND location IS NOT NULL
		  AND ST_DWithin(
			location,
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
			?
		  )
		ORDER BY ST_Distance(
			location,
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography
		)
	`, lng, lat, radiusMetres, lng, lat).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]geodomain.GeofenceZone, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out, nil
}

func (r *GeofenceRepository) Link(ctx context.Context, campaignID string, zoneIDs []string) error {
	if campaignID == "" {
		return fmt.Errorf("campaign_id is required")
	}
	rows := make([]persistance.CampaignGeofenceTargetRow, 0, len(zoneIDs))
	for _, zoneID := range zoneIDs {
		if zoneID == "" {
			continue
		}
		rows = append(rows, persistance.CampaignGeofenceTargetRow{
			CampaignID: campaignID,
			ZoneID:     zoneID,
		})
	}
	if len(rows) == 0 {
		return fmt.Errorf("at least one zone_id is required")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (r *GeofenceRepository) ListZonesForCampaign(ctx context.Context, campaignID string) ([]geodomain.GeofenceZone, error) {
	var rows []persistance.GeofenceZoneRow
	err := r.db.WithContext(ctx).
		Table("geofence_zones AS z").
		Select("z.*").
		Joins("JOIN campaign_geofence_targets t ON t.zone_id = z.id").
		Where("t.campaign_id = ?", campaignID).
		Order("z.created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]geodomain.GeofenceZone, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out, nil
}

type eligibleCampaignScan struct {
	campaignpersistence.CampaignRow
	PlanID            string  `gorm:"column:plan_id"`
	PlanName          string  `gorm:"column:plan_name"`
	PlanMonthlyFeeETB float64 `gorm:"column:plan_monthly_fee_etb"`
	ChannelCode       string  `gorm:"column:channel_code"`
}

func (r *GeofenceRepository) ListEligibleCampaignsForZone(
	ctx context.Context,
	zoneID, targetIntent string,
) ([]campaigndomain.Campaign, error) {
	now := time.Now().UTC()
	var rows []eligibleCampaignScan
	q := r.db.WithContext(ctx).
		Table("campaigns").
		Select(`campaigns.*,
			sp.id AS plan_id,
			sp.name AS plan_name,
			sp.monthly_fee_etb AS plan_monthly_fee_etb,
			channels.code AS channel_code`).
		Joins("JOIN campaign_geofence_targets t ON t.campaign_id = campaigns.id").
		Joins("JOIN channels ON channels.id = campaigns.channel_id").
		Joins(`JOIN advertiser_subscriptions sub ON sub.advertiser_id = campaigns.advertiser_id AND sub.status = 'active'`).
		Joins(`JOIN subscription_plans sp ON sp.id = sub.plan_id AND sp.is_active = true`).
		Where(`t.zone_id = ?
			AND campaigns.is_active = ?
			AND campaigns.validation_status = ?
			AND campaigns.moderation_status = ?
			AND sub.current_period_start <= ? AND sub.current_period_end >= ?`,
			zoneID, true, "passed", campaigndomain.ModerationApproved, now, now)
	if targetIntent != "" {
		q = q.Where("campaigns.target_intent = ?", targetIntent)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]campaigndomain.Campaign, 0, len(rows))
	for _, row := range rows {
		c, err := row.CampaignRow.ToDomain()
		if err != nil {
			return nil, err
		}
		c.PlanID = row.PlanID
		c.PlanName = row.PlanName
		c.PlanMonthlyFeeETB = row.PlanMonthlyFeeETB
		c.ChannelCode = row.ChannelCode
		out = append(out, *c)
	}
	return out, nil
}

func (r *GeofenceRepository) HasLocationConsent(ctx context.Context, pseudonymousID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) FROM demo_sms_recipients
		WHERE pseudonymous_id = ? AND is_active = TRUE
	`, pseudonymousID).Scan(&count).Error; err != nil {
		return false, err
	}
	if count == 0 {
		return false, gorm.ErrRecordNotFound
	}
	var consented bool
	if err := r.db.WithContext(ctx).Raw(`
		SELECT location_ad_consent
		FROM demo_sms_recipients
		WHERE pseudonymous_id = ? AND is_active = TRUE
		LIMIT 1
	`, pseudonymousID).Scan(&consented).Error; err != nil {
		return false, err
	}
	return consented, nil
}

func (r *GeofenceRepository) SetLocationConsent(
	ctx context.Context,
	pseudonymousID string,
	consented bool,
) error {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE demo_sms_recipients
		SET location_ad_consent = ?, updated_at = NOW()
		WHERE pseudonymous_id = ? AND is_active = TRUE
	`, consented, pseudonymousID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// StoreVisitRepository persists zone-entry events.
type StoreVisitRepository struct {
	db *gorm.DB
}

func NewStoreVisitRepository(db *gorm.DB) *StoreVisitRepository {
	return &StoreVisitRepository{db: db}
}

var _ geodomain.VisitRepository = (*StoreVisitRepository)(nil)

func (r *StoreVisitRepository) Create(ctx context.Context, visit *geodomain.StoreVisit) error {
	if visit.ID == "" {
		visit.ID = uuid.NewString()
	}
	if visit.VisitedAt.IsZero() {
		visit.VisitedAt = time.Now().UTC()
	}
	row := persistance.StoreVisitRowFromDomain(*visit)
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	*visit = row.ToDomain()
	return nil
}

func (r *StoreVisitRepository) CountByUserExcluding(ctx context.Context, pseudonymousID, excludeVisitID string) (int64, error) {
	var count int64
	q := r.db.WithContext(ctx).Model(&persistance.StoreVisitRow{}).
		Where("pseudonymous_id = ?", pseudonymousID)
	if excludeVisitID != "" {
		q = q.Where("id <> ?", excludeVisitID)
	}
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
