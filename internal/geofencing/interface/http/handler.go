package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	adportaldomain "skykin-platform/internal/ad_portal/domain"
	geoapp "skykin-platform/internal/geofencing/application"
	geodomain "skykin-platform/internal/geofencing/domain"
	platformHTTP "skykin-platform/internal/platform/http"
	platformMiddleware "skykin-platform/internal/platform/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *geoapp.GeofencingService
}

func NewHandler(svc *geoapp.GeofencingService) *Handler {
	return &Handler{svc: svc}
}

// CreateZone godoc
// @Summary      Create geofence zone (store)
// @Description  Advertiser creates a store geofence. Zone is saved with is_active=false until an operator approves a campaign linked to it.
// @Tags         Ad Portal - Geofences
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  CreateZoneRequest  true  "Zone"
// @Success      201  {object}  ZoneDTO
// @Failure      400  {object}  platformHTTP.APIError
// @Router       /ad-portal/geofences [post]
func (h *Handler) CreateZone(c *gin.Context) {
	if !isAdvertiser(c) {
		platformHTTP.Error(c, http.StatusForbidden, "only advertisers can create geofence zones", nil)
		return
	}
	var req CreateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	advertiserID := platformMiddleware.AccountAdvertiserID(c)
	if advertiserID == "" {
		platformHTTP.Error(c, http.StatusForbidden, "only advertisers can create geofence zones", nil)
		return
	}
	zone, err := h.svc.CreateZone(c.Request.Context(), geoapp.CreateZoneCommand{
		AdvertiserID: advertiserID,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMetres: req.RadiusMetres,
	})
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "create geofence failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, mapZone(*zone, true))
}

// ListZones godoc
// @Summary      List advertiser geofence zones
// @Tags         Ad Portal - Geofences
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  ZoneDTO
// @Router       /ad-portal/geofences [get]
func (h *Handler) ListZones(c *gin.Context) {
	advertiserID := platformMiddleware.AccountAdvertiserID(c)
	zones, err := h.svc.ListZones(c.Request.Context(), advertiserID)
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list geofences failed", nil)
		return
	}
	out := make([]ZoneDTO, 0, len(zones))
	for _, z := range zones {
		out = append(out, mapZone(z, true))
	}
	c.JSON(http.StatusOK, out)
}

// LinkCampaignZones godoc
// @Summary      Link campaign to geofence zones
// @Tags         Ad Portal - Geofences
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string  true  "Campaign ID"
// @Param        body  body  LinkCampaignZonesRequest  true  "Zone IDs"
// @Success      204
// @Router       /ad-portal/campaigns/{id}/geofences [post]
func (h *Handler) LinkCampaignZones(c *gin.Context) {
	if !isAdvertiser(c) {
		platformHTTP.Error(c, http.StatusForbidden, "only advertisers can link geofence zones", nil)
		return
	}
	var req LinkCampaignZonesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	advertiserID := platformMiddleware.AccountAdvertiserID(c)
	if advertiserID == "" {
		platformHTTP.Error(c, http.StatusForbidden, "only advertisers can link geofence zones", nil)
		return
	}
	err := h.svc.LinkCampaignZones(c.Request.Context(), advertiserID, c.Param("id"), req.ZoneIDs)
	if err != nil {
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, geoapp.ErrAdvertiserOnly), errors.Is(err, geoapp.ErrCampaignNotOwned):
			status = http.StatusForbidden
		case errors.Is(err, geoapp.ErrZoneNotFound):
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "link geofences failed", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// ListPendingZones godoc
// @Summary      List inactive geofence zones pending activation
// @Description  Operator reviews advertiser-created stores that are not yet is_active=true.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ZoneListResponse
// @Router       /ad-portal/admin/geofences/pending [get]
func (h *Handler) ListPendingZones(c *gin.Context) {
	zones, err := h.svc.ListPendingZones(c.Request.Context())
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list pending geofences failed", err.Error())
		return
	}
	out := make([]ZoneDTO, 0, len(zones))
	for _, z := range zones {
		out = append(out, mapZone(z, true))
	}
	c.JSON(http.StatusOK, ZoneListResponse{Zones: out, Count: len(out)})
}

// ActivateZone godoc
// @Summary      Activate a geofence zone if not already active
// @Description  Operator approves a draft store. Idempotent when the zone is already active.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Zone ID"
// @Success      200  {object}  ZoneDTO
// @Failure      404  {object}  platformHTTP.APIError
// @Router       /ad-portal/admin/geofences/{id}/activate [post]
func (h *Handler) ActivateZone(c *gin.Context) {
	zone, err := h.svc.ActivateZone(c.Request.Context(), c.Param("id"))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, geoapp.ErrZoneNotFound) {
			status = http.StatusNotFound
		}
		platformHTTP.Error(c, status, "activate geofence failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, mapZone(*zone, true))
}

// ActivateCampaignZones godoc
// @Summary      Activate inactive geofence zones linked to a campaign
// @Description  Operator turns on draft stores attached to the campaign. Already-active zones are left unchanged.
// @Tags         Ad Portal - Admin
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Campaign ID"
// @Success      200  {object}  ZoneListResponse
// @Router       /ad-portal/admin/campaigns/{id}/geofences/activate [post]
func (h *Handler) ActivateCampaignZones(c *gin.Context) {
	zones, err := h.svc.ActivateCampaignZones(c.Request.Context(), c.Param("id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "activate campaign geofences failed", err.Error())
		return
	}
	out := make([]ZoneDTO, 0, len(zones))
	for _, z := range zones {
		out = append(out, mapZone(z, true))
	}
	c.JSON(http.StatusOK, ZoneListResponse{Zones: out, Count: len(out)})
}

// ListCampaignZones godoc
// @Summary      List zones linked to a campaign
// @Tags         Ad Portal - Geofences
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Campaign ID"
// @Success      200  {array}  ZoneDTO
// @Router       /ad-portal/campaigns/{id}/geofences [get]
func (h *Handler) ListCampaignZones(c *gin.Context) {
	zones, err := h.svc.ListCampaignZones(c.Request.Context(), c.Param("id"))
	if err != nil {
		platformHTTP.Error(c, http.StatusInternalServerError, "list campaign geofences failed", nil)
		return
	}
	out := make([]ZoneDTO, 0, len(zones))
	for _, z := range zones {
		out = append(out, mapZone(z, true))
	}
	c.JSON(http.StatusOK, out)
}

// SetLocationConsent godoc
// @Summary      Set demo location ad consent
// @Description  Updates location_ad_consent on demo_sms_recipients only.
// @Tags         SDK - Geofences
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  LocationConsentRequest  true  "Consent"
// @Success      200  {object}  LocationConsentResponse
// @Failure      404  {object}  platformHTTP.APIError
// @Router       /geofences/location-consent [patch]
func (h *Handler) SetLocationConsent(c *gin.Context) {
	var req LocationConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	err := h.svc.SetLocationConsent(c.Request.Context(), req.PseudonymousID, req.LocationAdConsent)
	if err != nil {
		if errors.Is(err, geoapp.ErrDemoRecipientNotFound) {
			platformHTTP.Error(c, http.StatusNotFound, "demo recipient not found", nil)
			return
		}
		platformHTTP.Error(c, http.StatusBadRequest, "update location consent failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, LocationConsentResponse{
		Status:            "success",
		PseudonymousID:    req.PseudonymousID,
		LocationAdConsent: req.LocationAdConsent,
	})
}

// Sync godoc
// @Summary      Sync nearby geofence zones
// @Description  Returns active zones within radius_m (default 20000) using PostGIS ST_DWithin.
// @Tags         SDK - Geofences
// @Produce      json
// @Security     APIKeyAuth
// @Param        lat       query  number  true   "Latitude"
// @Param        lng       query  number  true   "Longitude"
// @Param        radius_m  query  int     false  "Search radius metres" default(20000)
// @Success      200  {object}  SyncResponse
// @Router       /geofences/sync [get]
func (h *Handler) Sync(c *gin.Context) {
	lat, errLat := strconv.ParseFloat(c.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(c.Query("lng"), 64)
	if errLat != nil || errLng != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "lat and lng are required numbers", nil)
		return
	}
	radius := 20000
	if raw := c.Query("radius_m"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			platformHTTP.Error(c, http.StatusBadRequest, "radius_m must be an integer", nil)
			return
		}
		radius = parsed
	}
	zones, err := h.svc.SyncNearby(c.Request.Context(), geoapp.SyncQuery{
		Latitude: lat, Longitude: lng, RadiusMetres: radius,
	})
	if err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "geofence sync failed", err.Error())
		return
	}
	out := make([]ZoneDTO, 0, len(zones))
	for _, z := range zones {
		out = append(out, mapZone(z, false))
	}
	c.JSON(http.StatusOK, SyncResponse{Status: "success", Zones: out})
}

// Event godoc
// @Summary      Record geofence enter event and return ad creative
// @Tags         SDK - Geofences
// @Accept       json
// @Produce      json
// @Security     APIKeyAuth && SDKSecretAuth
// @Param        body  body  GeofenceEventRequest  true  "Enter event"
// @Success      202  {object}  GeofenceEventResponse
// @Failure      403  {object}  platformHTTP.APIError
// @Router       /geofence/event [post]
func (h *Handler) Event(c *gin.Context) {
	var req GeofenceEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		platformHTTP.Error(c, http.StatusBadRequest, "invalid payload", err.Error())
		return
	}
	result, err := h.svc.ProcessEvent(c.Request.Context(), geoapp.ProcessEventCommand{
		PseudonymousID: req.PseudonymousID,
		ZoneID:         req.ZoneID,
		AccuracyM:      req.AccuracyM,
	})
	if err != nil {
		status := http.StatusInternalServerError
		msg := "geofence event failed"
		switch {
		case errors.Is(err, geoapp.ErrLocationConsentDenied):
			status = http.StatusForbidden
			msg = "location ad consent required"
		case errors.Is(err, geoapp.ErrDemoRecipientNotFound):
			status = http.StatusNotFound
			msg = "demo recipient not found"
		case errors.Is(err, geoapp.ErrZoneNotFound):
			status = http.StatusNotFound
			msg = "zone not found"
		case errors.Is(err, geoapp.ErrInvalidGeofence):
			status = http.StatusBadRequest
			msg = "invalid geofence event"
		}
		platformHTTP.Error(c, status, msg, err.Error())
		return
	}
	resp := GeofenceEventResponse{
		Status:    "accepted",
		VisitID:   result.VisitID,
		VisitedAt: result.VisitedAt.UTC().Format(time.RFC3339Nano),
	}
	if result.AdReturned {
		resp.CampaignID = result.CampaignID
		resp.CampaignName = result.CampaignName
		resp.ChannelCode = result.ChannelCode
		resp.AdContent = result.AdContent
	}
	c.JSON(http.StatusAccepted, resp)
}

func mapZone(z geodomain.GeofenceZone, includeMeta bool) ZoneDTO {
	dto := ZoneDTO{
		ID:           z.ID,
		Latitude:     z.Latitude,
		Longitude:    z.Longitude,
		RadiusMetres: z.RadiusMetres,
		IsActive:     z.IsActive,
	}
	if includeMeta {
		dto.AdvertiserID = z.AdvertiserID
		dto.CreatedAt = z.CreatedAt.UTC().Format(time.RFC3339Nano)
		dto.UpdatedAt = z.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	return dto
}

func isAdvertiser(c *gin.Context) bool {
	role, _ := c.Get("portal_role")
	roleStr, _ := role.(string)
	return roleStr == adportaldomain.RoleAdvertiser
}
