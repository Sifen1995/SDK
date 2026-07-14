package http

// CreateConsentRequest is the Flutter SDK payload for consent registration.
type CreateConsentRequest struct {
	// ConsentLevel controls identity reversibility: individual | aggregate | none
	ConsentLevel string `json:"consent_level" binding:"required" example:"individual" enums:"individual,aggregate,none"`
	// SDKVersion is the Flutter SDK version string
	SDKVersion string `json:"sdk_version" binding:"required" example:"1.0.0"`
}

// CreateConsentResponse is returned after a successful registration.
type CreateConsentResponse struct {
	// Status is always "success" on HTTP 201
	Status string `json:"status" example:"success"`
	// ConsentLevel echoes the granted level
	ConsentLevel string `json:"consent_level" example:"individual"`
	// PseudonymousID is a backend-generated UUID for Flutter to attach to intents
	PseudonymousID string `json:"pseudonymous_id" example:"9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"`
}
