package http

import "time"

// IngestEventsRequest is the SDK batch ingestion payload. Events are keyed by the
// consent-issued pseudonymous id; user_id is accepted as the legacy alias.
type IngestEventsRequest struct {
	PseudonymousID string       `json:"pseudonymous_id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	UserID         string       `json:"user_id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	Events         []EventInput `json:"events" binding:"required,min=1,dive"`
}

// SubjectID returns the pseudonymous id regardless of which alias the client sent.
func (r IngestEventsRequest) SubjectID() string {
	if r.PseudonymousID != "" {
		return r.PseudonymousID
	}
	return r.UserID
}

// EventInput is a single SDK event in a batch.
type EventInput struct {
	EventID    string                 `json:"event_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	EventType  string                 `json:"event_type" binding:"required" example:"content_viewed" enums:"session_started,screen_viewed,content_viewed,search_performed,interaction_received,scroll_activity,notification_opened,campaign_impression,campaign_clicked,conversion_completed,transaction_completed,reward_claimed"`
	Domain     string                 `json:"domain" binding:"required" example:"crypto"`
	SessionID  string                 `json:"session_id" example:"660e8400-e29b-41d4-a716-446655440002"`
	ScreenName string                 `json:"screen_name" example:"asset_details"`
	Metadata   map[string]interface{} `json:"metadata" binding:"required"`
	DeviceType string                 `json:"device_type" example:"mobile"`
	Platform   string                 `json:"platform" example:"android"`
	AppVersion string                 `json:"app_version" example:"1.2.0"`
	CreatedAt  *time.Time             `json:"created_at,omitempty" example:"2026-06-01T12:00:00Z"`
}

// IngestEventsResponse is returned with HTTP 202 Accepted.
type IngestEventsResponse struct {
	Accepted         bool                   `json:"accepted"`
	PredictionQueued bool                   `json:"prediction_queued"`
	Results          []EventIngestResultDTO `json:"results"`
}

// EventIngestResultDTO describes per-event ingestion status.
type EventIngestResultDTO struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}
