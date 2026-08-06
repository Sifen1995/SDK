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

type stubTargets struct {
	campaigns []campaigndomain.Campaign
}

func (stubTargets) Link(context.Context, string, []string) error { return nil }
func (stubTargets) ListZonesForCampaign(context.Context, string) ([]geodomain.GeofenceZone, error) {
	return nil, nil
}
func (s stubTargets) ListEligibleCampaignsForZone(context.Context, string) ([]campaigndomain.Campaign, error) {
	return s.campaigns, nil
}

type stubVisits struct{}

func (stubVisits) Create(_ context.Context, visit *geodomain.StoreVisit) error {
	visit.ID = "visit-1"
	return nil
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

func TestProcessEventRequiresConsent(t *testing.T) {
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		stubTargets{},
		stubVisits{},
		stubConsent{consented: false},
		nil, nil, nil,
	)
	_, err := svc.ProcessEvent(context.Background(), ProcessEventCommand{
		PseudonymousID: "p1", ZoneID: "z1",
	})
	require.ErrorIs(t, err, ErrLocationConsentDenied)
}

func TestProcessEventSkipsCreativeWhenCapped(t *testing.T) {
	svc := NewGeofencingService(
		stubZones{zone: &geodomain.GeofenceZone{ID: "z1"}},
		stubTargets{campaigns: []campaigndomain.Campaign{{
			ID: "c1", Name: "Macchiato", Title: "20% off", BodyText: "Today only",
			FrequencyCapPerDay: 3, ChannelCode: "PUSH", TotalBudgetCap: 100,
		}}},
		stubVisits{},
		stubConsent{consented: true},
		stubFreq{capped: true},
		nil, nil,
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
		stubTargets{campaigns: []campaigndomain.Campaign{{
			ID: "c1", Name: "Macchiato", Title: "20% off", BodyText: "Today only",
			ImageURL: "https://img", DestinationURL: "https://kaldi",
			FrequencyCapPerDay: 3, ChannelCode: "PUSH", TotalBudgetCap: 100,
		}}},
		stubVisits{},
		stubConsent{consented: true},
		stubFreq{capped: false},
		nil, nil,
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
	svc := NewGeofencingService(nil, nil, nil, stubConsent{err: gorm.ErrRecordNotFound}, nil, nil, nil)
	err := svc.SetLocationConsent(context.Background(), "missing", true)
	require.ErrorIs(t, err, ErrDemoRecipientNotFound)
}

func TestCreateZoneValidatesCoordinates(t *testing.T) {
	svc := NewGeofencingService(stubZones{zone: &geodomain.GeofenceZone{}}, nil, nil, nil, nil, nil, nil)
	_, err := svc.CreateZone(context.Background(), CreateZoneCommand{
		AdvertiserID: "a1", Latitude: 200, Longitude: 0, RadiusMetres: 150,
	})
	require.True(t, errors.Is(err, ErrInvalidGeofence))
}
