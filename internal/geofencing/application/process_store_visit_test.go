package application

import (
	"context"
	"errors"
	"testing"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	geodomain "skykin-platform/internal/geofencing/domain"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubZones struct {
	zone *geodomain.GeofenceZone
}

func (s stubZones) Create(_ context.Context, zone *geodomain.GeofenceZone) error {
	*zone = *s.zone
	return nil
}
func (s stubZones) ListByAdvertiser(context.Context, string) ([]geodomain.GeofenceZone, error) {
	return []geodomain.GeofenceZone{*s.zone}, nil
}
func (s stubZones) ListInactive(context.Context) ([]geodomain.GeofenceZone, error) {
	return []geodomain.GeofenceZone{*s.zone}, nil
}
func (s stubZones) GetByID(context.Context, string) (*geodomain.GeofenceZone, error) {
	return s.zone, nil
}
func (s stubZones) FindNearby(context.Context, float64, float64, int) ([]geodomain.GeofenceZone, error) {
	return []geodomain.GeofenceZone{*s.zone}, nil
}
func (s stubZones) ActivateZone(context.Context, string) (*geodomain.GeofenceZone, error) {
	z := *s.zone
	z.IsActive = true
	return &z, nil
}
func (stubZones) ActivateForCampaign(context.Context, string) error { return nil }

// stubTargets returns intent-filtered campaigns when targetIntent is set,
// and allCampaigns (or campaigns) when the visit-history path uses "".
type stubTargets struct {
	byIntent     map[string][]campaigndomain.Campaign
	allCampaigns []campaigndomain.Campaign
	lastIntent   string
}

func (stubTargets) Link(context.Context, string, []string) error { return nil }
func (stubTargets) ListZonesForCampaign(context.Context, string) ([]geodomain.GeofenceZone, error) {
	return nil, nil
}
func (s *stubTargets) ListEligibleCampaignsForZone(_ context.Context, _ string, targetIntent string) ([]campaigndomain.Campaign, error) {
	s.lastIntent = targetIntent
	if targetIntent == "" {
		if s.allCampaigns != nil {
			return s.allCampaigns, nil
		}
		return nil, nil
	}
	if s.byIntent != nil {
		return s.byIntent[targetIntent], nil
	}
	return nil, nil
}

type stubVisits struct {
	priorCount int64
}

func (stubVisits) Create(_ context.Context, visit *geodomain.StoreVisit) error {
	visit.ID = "visit-1"
	return nil
}
func (s stubVisits) CountByUserExcluding(context.Context, string, string) (int64, error) {
	return s.priorCount, nil
}

type stubConsent struct {
	consented bool
	err       error
}

func (s stubConsent) HasLocationConsent(context.Context, string) (bool, error) {
	return s.consented, s.err
}
func (s stubConsent) SetLocationConsent(context.Context, string, bool) error { return s.err }

type stubFreq struct {
	capped bool
}

func (s stubFreq) IsFrequencyCapped(context.Context, string, string, int) (bool, error) {
	return s.capped, nil
}
func (stubFreq) IncrementDeliveryCount(context.Context, string, string, time.Duration) (int64, error) {
	return 1, nil
}

type stubIntent struct {
	name string
}

func (s stubIntent) CurrentIntent(context.Context, string) (string, error) {
	return s.name, nil
}

func sampleCampaign(id, name string) campaigndomain.Campaign {
	return campaigndomain.Campaign{
		ID: id, Name: name, Title: "20% off", BodyText: "Today only",
		ImageURL: "https://img", DestinationURL: "https://kaldi",
		FrequencyCapPerDay: 3, ChannelCode: "PUSH", TotalBudgetCap: 100,
	}
}

func TestProcessEventRequiresConsent(t *testing.T) {
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		&stubTargets{},
		stubVisits{},
		stubConsent{consented: false},
		nil, nil, nil, nil,
	)
	_, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.ErrorIs(t, err, ErrLocationConsentDenied)
}

func TestProcessEventFirstVisitNoIntentNoAd(t *testing.T) {
	targets := &stubTargets{
		allCampaigns: []campaigndomain.Campaign{sampleCampaign("c1", "Macchiato")},
	}
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		targets,
		stubVisits{priorCount: 0},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
		stubIntent{name: ""},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.Equal(t, "visit-1", result.VisitID)
	require.False(t, result.AdReturned)
}

func TestProcessEventFirstVisitMatchingIntentReturnsAd(t *testing.T) {
	targets := &stubTargets{
		byIntent: map[string][]campaigndomain.Campaign{
			"coffee_interest": {sampleCampaign("c-intent", "Coffee")},
		},
		allCampaigns: []campaigndomain.Campaign{sampleCampaign("c-all", "All")},
	}
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		targets,
		stubVisits{priorCount: 0},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
		stubIntent{name: "coffee_interest"},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.True(t, result.AdReturned)
	require.Equal(t, "c-intent", result.CampaignID)
	require.Equal(t, "coffee_interest", targets.lastIntent)
}

func TestProcessEventPriorVisitsNoIntentReturnsAd(t *testing.T) {
	targets := &stubTargets{
		allCampaigns: []campaigndomain.Campaign{sampleCampaign("c-zone", "Zone")},
	}
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		targets,
		stubVisits{priorCount: 2},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
		stubIntent{name: ""},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.True(t, result.AdReturned)
	require.Equal(t, "c-zone", result.CampaignID)
	require.Equal(t, "", targets.lastIntent)
}

func TestProcessEventPriorVisitsIntentMatchPrefersFiltered(t *testing.T) {
	targets := &stubTargets{
		byIntent: map[string][]campaigndomain.Campaign{
			"coffee_interest": {sampleCampaign("c-intent", "Coffee")},
		},
		allCampaigns: []campaigndomain.Campaign{sampleCampaign("c-all", "All")},
	}
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		targets,
		stubVisits{priorCount: 1},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
		stubIntent{name: "coffee_interest"},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.True(t, result.AdReturned)
	require.Equal(t, "c-intent", result.CampaignID)
	require.Equal(t, "coffee_interest", targets.lastIntent)
}

func TestProcessEventPriorVisitsIntentMismatchFallsBackToHistory(t *testing.T) {
	targets := &stubTargets{
		byIntent: map[string][]campaigndomain.Campaign{
			"coffee_interest": {sampleCampaign("c-intent", "Coffee")},
		},
		allCampaigns: []campaigndomain.Campaign{sampleCampaign("c-all", "All")},
	}
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		targets,
		stubVisits{priorCount: 1},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
		stubIntent{name: "unrelated_intent"},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.True(t, result.AdReturned)
	require.Equal(t, "c-all", result.CampaignID)
	require.Equal(t, "", targets.lastIntent)
}

func TestProcessEventSkipsCreativeWhenCapped(t *testing.T) {
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		&stubTargets{
			allCampaigns: []campaigndomain.Campaign{sampleCampaign("c1", "Macchiato")},
		},
		stubVisits{priorCount: 1},
		stubConsent{consented: true},
		stubFreq{capped: true},
		nil, nil,
		stubIntent{},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.Equal(t, "visit-1", result.VisitID)
	require.False(t, result.AdReturned)
}

func TestProcessEventReturnsCreative(t *testing.T) {
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		&stubTargets{
			byIntent: map[string][]campaigndomain.Campaign{
				"coffee_interest": {sampleCampaign("c1", "Macchiato")},
			},
		},
		stubVisits{priorCount: 0},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
		stubIntent{name: "coffee_interest"},
	)
	result, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.NoError(t, err)
	require.True(t, result.AdReturned)
	require.Equal(t, "c1", result.CampaignID)
	require.Equal(t, "20% off", result.AdContent["title"])
}

func TestSetLocationConsentMapsMissingRecipient(t *testing.T) {
	svc := NewGeofencingService(nil, nil, nil, stubConsent{err: gorm.ErrRecordNotFound}, nil, nil, nil, nil)
	err := svc.SetLocationConsent(context.Background(), "missing", true)
	require.ErrorIs(t, err, ErrDemoRecipientNotFound)
}

func TestCreateZoneValidatesCoordinates(t *testing.T) {
	svc := NewGeofencingService(stubZones{zone: &geodomain.GeofenceZone{}}, nil, nil, nil, nil, nil, nil, nil)
	_, err := svc.CreateZone(context.Background(), CreateZoneCommand{
		AdvertiserID: "a1", Latitude: 200, Longitude: 0, RadiusMetres: 150,
	})
	require.True(t, errors.Is(err, ErrInvalidGeofence))
}
