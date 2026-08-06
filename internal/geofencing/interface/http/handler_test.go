package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	geoapp "skykin-platform/internal/geofencing/application"
	geodomain "skykin-platform/internal/geofencing/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type zoneStub struct{}

func (zoneStub) Create(_ context.Context, zone *geodomain.GeofenceZone) error {
	zone.ID = "zone-1"
	return nil
}
func (zoneStub) ListByAdvertiser(context.Context, string) ([]geodomain.GeofenceZone, error) {
	return nil, nil
}
func (zoneStub) ListInactive(context.Context) ([]geodomain.GeofenceZone, error) {
	return nil, nil
}
func (zoneStub) GetByID(context.Context, string) (*geodomain.GeofenceZone, error) {
	return &geodomain.GeofenceZone{ID: "zone-1", IsActive: true}, nil
}
func (zoneStub) FindNearby(context.Context, float64, float64, int) ([]geodomain.GeofenceZone, error) {
	return []geodomain.GeofenceZone{{
		ID: "zone-1", Latitude: 9.0227, Longitude: 38.7468, RadiusMetres: 150, IsActive: true,
	}}, nil
}
func (zoneStub) ActivateZone(_ context.Context, zoneID string) (*geodomain.GeofenceZone, error) {
	return &geodomain.GeofenceZone{ID: zoneID, IsActive: true}, nil
}
func (zoneStub) ActivateForCampaign(context.Context, string) error { return nil }

type nilTargets struct{}

func (nilTargets) Link(context.Context, string, []string) error { return nil }
func (nilTargets) ListZonesForCampaign(context.Context, string) ([]geodomain.GeofenceZone, error) {
	return nil, nil
}
func (nilTargets) ListEligibleCampaignsForZone(context.Context, string) ([]campaigndomain.Campaign, error) {
	return nil, nil
}

type visitOK struct{}

func (visitOK) Create(_ context.Context, visit *geodomain.StoreVisit) error {
	visit.ID = "v1"
	visit.VisitedAt = time.Now().UTC()
	return nil
}

type consentDenied struct{}

func (consentDenied) HasLocationConsent(context.Context, string) (bool, error) { return false, nil }
func (consentDenied) SetLocationConsent(context.Context, string, bool) error   { return nil }

func TestSyncHandlerRequiresLatLng(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := geoapp.NewGeofencingService(zoneStub{}, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/geofences/sync", NewHandler(svc).Sync)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/geofences/sync", nil))
	require.Equal(t, http.StatusBadRequest, w.Code)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/geofences/sync?lat=9.02&lng=38.74", nil))
	require.Equal(t, http.StatusOK, w2.Code)
	var body SyncResponse
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &body))
	require.Equal(t, "success", body.Status)
	require.Len(t, body.Zones, 1)
}

func TestEventHandlerForbiddenWithoutConsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := geoapp.NewGeofencingService(
		zoneStub{}, nilTargets{}, visitOK{}, consentDenied{}, nil, nil, nil,
	)
	router := gin.New()
	router.POST("/geofence/event", NewHandler(svc).Event)
	body := `{"pseudonymous_id":"p1","zone_id":"zone-1"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/geofence/event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}
