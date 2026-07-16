package http

// IngestIntentAdRequest is sent by Flutter after on-device ML prediction.
type IngestIntentAdRequest struct {
	// PseudonymousID from POST /consent (UUID)
	PseudonymousID string `json:"pseudonymous_id" binding:"required" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	// IntentName is the predicted intent class (e.g. fashion_interest)
	IntentName string `json:"intent_name" binding:"required" example:"fashion_interest"`
	// Confidence is the model score between 0 and 1
	Confidence float64 `json:"confidence" binding:"required" example:"0.87"`
	// ModelVersion identifies the on-device / ML model version
	ModelVersion string `json:"model_version" binding:"omitempty" example:"1.0.0"`
	// ChannelCode delivery channel; empty tries IN_APP_BANNER, SMS_PLUS, PUSH, NATIVE_FEED
	ChannelCode string `json:"channel_code" binding:"omitempty" example:"IN_APP_BANNER"`
}

// IngestIntentAdResponse returns the persisted intent and matched campaign creative.
type IngestIntentAdResponse struct {
	PseudonymousID string         `json:"pseudonymous_id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
	IntentName     string         `json:"intent_name" example:"fashion_interest"`
	Confidence     float64        `json:"confidence" example:"0.87"`
	ModelVersion   string         `json:"model_version" example:"1.0.0"`
	CampaignID     string         `json:"campaign_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CampaignName   string         `json:"campaign_name" example:"Summer Fashion Drop"`
	ChannelCode    string         `json:"channel_code" example:"IN_APP_BANNER"`
	AdContent      map[string]any `json:"ad_content"`
}

// IngestIntentAggregateRequest is a batch of anonymized intent signals from one device.
type IngestIntentAggregateRequest struct {
	// DateBucket is YYYY-MM-DD for the rollup day
	DateBucket string `json:"date_bucket" binding:"required" example:"2026-07-16"`
	// Intents is the per-intent counters for this device batch
	Intents []IngestIntentAggregateItem `json:"intents" binding:"required,min=1,dive"`
}

// IngestIntentAggregateItem is one intent entry inside the device batch.
type IngestIntentAggregateItem struct {
	IntentName     string  `json:"intent_name" binding:"required" example:"fashion_interest"`
	Count          int     `json:"count" binding:"required,min=1" example:"3"`
	DaysConsistent float64 `json:"days_consistent" binding:"required,min=1" example:"7"`
}
