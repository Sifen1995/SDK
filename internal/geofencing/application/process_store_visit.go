package application

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	campaignApp "skykin-platform/internal/campaigns/application"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	geodomain "skykin-platform/internal/geofencing/domain"

	"gorm.io/gorm"
)

var (
	ErrLocationConsentDenied = errors.New("location ad consent required")
	ErrInvalidGeofence       = errors.New("invalid geofence request")
	ErrZoneNotFound          = errors.New("geofence zone not found")
	ErrDemoRecipientNotFound = errors.New("demo recipient not found")
	ErrCampaignNotOwned      = errors.New("campaign not owned by advertiser")
	ErrAdvertiserOnly        = errors.New("only advertisers can create or link geofence zones")
)

type CreateZoneCommand struct {
	AdvertiserID string
	Latitude     float64
	Longitude    float64
	RadiusMetres int
}

type SyncQuery struct {
	Latitude     float64
	Longitude    float64
	RadiusMetres int
}

type ProcessEventCommand struct {
	PseudonymousID string
	ZoneID         string
	AccuracyM      *float64
}

type ProcessEventResult struct {
	VisitID      string
	VisitedAt    time.Time
	CampaignID   string
	CampaignName string
	ChannelCode  string
	AdContent    map[string]any
	AdReturned   bool
}

type CampaignOwnerChecker interface {
	GetByID(ctx context.Context, id string) (*campaigndomain.Campaign, error)
}

// ActiveIntentReader resolves the user's current intent for geofence ad matching.
type ActiveIntentReader interface {
	CurrentIntent(ctx context.Context, pseudonymousID string) (string, error)
}

type GeofencingService struct {
	zones     geodomain.ZoneRepository
	targets   geodomain.TargetRepository
	visits    geodomain.VisitRepository
	consent   geodomain.LocationConsentRepository
	freq      geodomain.FrequencyCapPort
	budget    geodomain.BudgetExhaustedPort
	campaigns CampaignOwnerChecker
	intents   ActiveIntentReader
}

func NewGeofencingService(
	zones geodomain.ZoneRepository,
	targets geodomain.TargetRepository,
	visits geodomain.VisitRepository,
	consent geodomain.LocationConsentRepository,
	freq geodomain.FrequencyCapPort,
	budget geodomain.BudgetExhaustedPort,
	campaigns CampaignOwnerChecker,
	intents ActiveIntentReader,
) *GeofencingService {
	return &GeofencingService{
		zones:     zones,
		targets:   targets,
		visits:    visits,
		consent:   consent,
		freq:      freq,
		budget:    budget,
		campaigns: campaigns,
		intents:   intents,
	}
}

func (s *GeofencingService) CreateZone(ctx context.Context, cmd CreateZoneCommand) (*geodomain.GeofenceZone, error) {
	if cmd.AdvertiserID == "" {
		return nil, fmt.Errorf("%w: advertiser_id required", ErrInvalidGeofence)
	}
	if cmd.Latitude < -90 || cmd.Latitude > 90 || cmd.Longitude < -180 || cmd.Longitude > 180 {
		return nil, fmt.Errorf("%w: invalid coordinates", ErrInvalidGeofence)
	}
	radius := cmd.RadiusMetres
	if radius <= 0 {
		radius = 150
	}
	if radius > 50000 {
		return nil, fmt.Errorf("%w: radius_metres too large", ErrInvalidGeofence)
	}
	// Advertiser drafts stay inactive until an operator activates them (alone or via campaign approve).
	zone := &geodomain.GeofenceZone{
		AdvertiserID: cmd.AdvertiserID,
		Latitude:     cmd.Latitude,
		Longitude:    cmd.Longitude,
		RadiusMetres: radius,
		IsActive:     false,
	}
	if err := s.zones.Create(ctx, zone); err != nil {
		return nil, err
	}
	return zone, nil
}

func (s *GeofencingService) ListZones(ctx context.Context, advertiserID string) ([]geodomain.GeofenceZone, error) {
	return s.zones.ListByAdvertiser(ctx, advertiserID)
}

func (s *GeofencingService) ListPendingZones(ctx context.Context) ([]geodomain.GeofenceZone, error) {
	return s.zones.ListInactive(ctx)
}

func (s *GeofencingService) ActivateZone(ctx context.Context, zoneID string) (*geodomain.GeofenceZone, error) {
	if zoneID == "" {
		return nil, fmt.Errorf("%w: zone_id required", ErrInvalidGeofence)
	}
	zone, err := s.zones.ActivateZone(ctx, zoneID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}
	return zone, nil
}

func (s *GeofencingService) ActivateCampaignZones(ctx context.Context, campaignID string) ([]geodomain.GeofenceZone, error) {
	if campaignID == "" {
		return nil, fmt.Errorf("%w: campaign_id required", ErrInvalidGeofence)
	}
	if err := s.zones.ActivateForCampaign(ctx, campaignID); err != nil {
		return nil, err
	}
	return s.targets.ListZonesForCampaign(ctx, campaignID)
}

func (s *GeofencingService) LinkCampaignZones(
	ctx context.Context,
	advertiserID, campaignID string,
	zoneIDs []string,
) error {
	if advertiserID == "" {
		return ErrAdvertiserOnly
	}
	if s.campaigns == nil {
		return fmt.Errorf("campaign ownership checker not configured")
	}
	campaign, err := s.campaigns.GetByID(ctx, campaignID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: campaign", ErrZoneNotFound)
		}
		return err
	}
	if campaign.AdvertiserID != advertiserID {
		return ErrCampaignNotOwned
	}
	for _, zoneID := range zoneIDs {
		zone, err := s.zones.GetByID(ctx, zoneID)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrZoneNotFound, zoneID)
		}
		if zone.AdvertiserID != advertiserID {
			return ErrCampaignNotOwned
		}
	}
	return s.targets.Link(ctx, campaignID, zoneIDs)
}

func (s *GeofencingService) ListCampaignZones(ctx context.Context, campaignID string) ([]geodomain.GeofenceZone, error) {
	return s.targets.ListZonesForCampaign(ctx, campaignID)
}

func (s *GeofencingService) SetLocationConsent(ctx context.Context, pseudonymousID string, consented bool) error {
	if pseudonymousID == "" {
		return fmt.Errorf("%w: pseudonymous_id required", ErrInvalidGeofence)
	}
	err := s.consent.SetLocationConsent(ctx, pseudonymousID, consented)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrDemoRecipientNotFound
	}
	return err
}

func (s *GeofencingService) SyncNearby(ctx context.Context, query SyncQuery) ([]geodomain.GeofenceZone, error) {
	if query.Latitude < -90 || query.Latitude > 90 || query.Longitude < -180 || query.Longitude > 180 {
		return nil, fmt.Errorf("%w: invalid coordinates", ErrInvalidGeofence)
	}
	radius := query.RadiusMetres
	if radius <= 0 {
		radius = 20000
	}
	if radius > 100000 {
		return nil, fmt.Errorf("%w: radius_m too large", ErrInvalidGeofence)
	}
	return s.zones.FindNearby(ctx, query.Latitude, query.Longitude, radius)
}

func (s *GeofencingService) ProcessEvent(
	ctx context.Context,
	cmd ProcessEventCommand,
) (*ProcessEventResult, error) {
	if cmd.PseudonymousID == "" || cmd.ZoneID == "" {
		return nil, fmt.Errorf("%w: pseudonymous_id and zone_id required", ErrInvalidGeofence)
	}
	consented, err := s.consent.HasLocationConsent(ctx, cmd.PseudonymousID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDemoRecipientNotFound
	}
	if err != nil {
		return nil, err
	}
	if !consented {
		return nil, ErrLocationConsentDenied
	}
	if _, err := s.zones.GetByID(ctx, cmd.ZoneID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}

	visit := &geodomain.StoreVisit{
		PseudonymousID: cmd.PseudonymousID,
		ZoneID:         cmd.ZoneID,
		AccuracyM:      cmd.AccuracyM,
		VisitedAt:      time.Now().UTC(),
	}
	if err := s.visits.Create(ctx, visit); err != nil {
		return nil, fmt.Errorf("record store visit: %w", err)
	}

	result := &ProcessEventResult{
		VisitID:   visit.ID,
		VisitedAt: visit.VisitedAt,
	}

	intentName := ""
	if s.intents != nil {
		name, err := s.intents.CurrentIntent(ctx, cmd.PseudonymousID)
		if err != nil {
			return result, err
		}
		intentName = name
	}

	priorCount, err := s.visits.CountByUserExcluding(ctx, cmd.PseudonymousID, visit.ID)
	if err != nil {
		return result, err
	}
	hasPriorVisits := priorCount > 0

	var campaigns []campaigndomain.Campaign
	if intentName != "" {
		campaigns, err = s.targets.ListEligibleCampaignsForZone(ctx, cmd.ZoneID, intentName)
		if err != nil {
			return result, err
		}
	}
	if len(campaigns) == 0 && hasPriorVisits {
		campaigns, err = s.targets.ListEligibleCampaignsForZone(ctx, cmd.ZoneID, "")
		if err != nil {
			return result, err
		}
	}
	if len(campaigns) == 0 {
		return result, nil
	}

	selected := s.pickCampaign(ctx, cmd.PseudonymousID, campaigns)
	if selected == nil {
		return result, nil
	}

	content, err := campaignApp.CampaignAdContent(selected, selected.ChannelCode)
	if err != nil {
		return result, err
	}
	if s.freq != nil {
		_, _ = s.freq.IncrementDeliveryCount(ctx, cmd.PseudonymousID, selected.ID, 24*time.Hour)
	}
	result.CampaignID = selected.ID
	result.CampaignName = selected.Name
	result.ChannelCode = selected.ChannelCode
	result.AdContent = content
	result.AdReturned = true
	return result, nil
}

func (s *GeofencingService) pickCampaign(
	ctx context.Context,
	pseudonymousID string,
	campaigns []campaigndomain.Campaign,
) *campaigndomain.Campaign {
	filtered := make([]campaigndomain.Campaign, 0, len(campaigns))
	for _, campaign := range campaigns {
		if campaign.BudgetSpent >= campaign.TotalBudgetCap && campaign.TotalBudgetCap > 0 {
			continue
		}
		if s.budget != nil {
			exhausted, err := s.budget.IsBudgetExhausted(ctx, campaign.ID)
			if err != nil || exhausted {
				continue
			}
		}
		capLimit := campaign.FrequencyCapPerDay
		if capLimit <= 0 {
			capLimit = 3
		}
		if s.freq != nil {
			capped, err := s.freq.IsFrequencyCapped(ctx, pseudonymousID, campaign.ID, capLimit)
			if err != nil || capped {
				continue
			}
		}
		filtered = append(filtered, campaign)
	}
	if len(filtered) == 0 {
		return nil
	}
	slices.SortFunc(filtered, func(a, b campaigndomain.Campaign) int {
		if a.PlanMonthlyFeeETB > b.PlanMonthlyFeeETB {
			return -1
		}
		if a.PlanMonthlyFeeETB < b.PlanMonthlyFeeETB {
			return 1
		}
		return 0
	})
	return &filtered[0]
}

// ProcessStoreVisit is retained as an alias type for the stub file name.
type ProcessStoreVisit = GeofencingService
