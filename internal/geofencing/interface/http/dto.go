package http

type CreateZoneRequest struct {
	Latitude     float64 `json:"latitude" binding:"required" example:"9.022736"`
	Longitude    float64 `json:"longitude" binding:"required" example:"38.746799"`
	RadiusMetres int     `json:"radius_metres" example:"150"`
}

type ZoneDTO struct {
	ID           string  `json:"id" example:"8ae89c1c-bfd8-40d2-9b78-50f5c46c413f"`
	AdvertiserID string  `json:"advertiser_id,omitempty"`
	Latitude     float64 `json:"latitude" example:"9.022736"`
	Longitude    float64 `json:"longitude" example:"38.746799"`
	RadiusMetres int     `json:"radius_metres" example:"150"`
	IsActive     bool    `json:"is_active" example:"true"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
}

type LinkCampaignZonesRequest struct {
	ZoneIDs []string `json:"zone_ids" binding:"required,min=1"`
}

type ZoneListResponse struct {
	Zones []ZoneDTO `json:"zones"`
	Count int       `json:"count"`
}

type LocationConsentRequest struct {
	PseudonymousID    string `json:"pseudonymous_id" binding:"required"`
	LocationAdConsent bool   `json:"location_ad_consent"`
}

type LocationConsentResponse struct {
	Status            string `json:"status" example:"success"`
	PseudonymousID    string `json:"pseudonymous_id"`
	LocationAdConsent bool   `json:"location_ad_consent"`
}

type SyncResponse struct {
	Status string    `json:"status" example:"success"`
	Zones  []ZoneDTO `json:"zones"`
}

type GeofenceEventRequest struct {
	PseudonymousID string   `json:"pseudonymous_id" binding:"required"`
	ZoneID         string   `json:"zone_id" binding:"required"`
	AccuracyM      *float64 `json:"accuracy_m,omitempty"`
}

type GeofenceEventResponse struct {
	Status       string         `json:"status" example:"accepted"`
	VisitID      string         `json:"visit_id"`
	VisitedAt    string         `json:"visited_at"`
	CampaignID   string         `json:"campaign_id,omitempty"`
	CampaignName string         `json:"campaign_name,omitempty"`
	ChannelCode  string         `json:"channel_code,omitempty"`
	AdContent    map[string]any `json:"ad_content,omitempty"`
}
