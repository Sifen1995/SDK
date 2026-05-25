package dto

import "time"

// TrackUserRequest handles incoming mapping identifiers from external apps
type TrackUserRequest struct {
	ExternalUserID string `json:"external_user_id" binding:"required"`
}

// TrackUserResponse returns the synced tracking state
type TrackUserResponse struct {
	ID             string    `json:"id"`
	ExternalUserID string    `json:"external_user_id"`
	CreatedAt      time.Time `json:"created_at"`
}
