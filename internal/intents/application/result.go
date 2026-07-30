package application

// IngestAndFetchAdResult is returned after persisting a profile and selecting an ad.
type IngestAndFetchAdResult struct {
	PseudonymousID string
	IntentName     string
	Confidence     float64
	ModelVersion   string
	CampaignID     string
	CampaignName   string
	ChannelCode    string
	AdContent      map[string]any
	// SMSDispatched is true when an SMS+ campaign was matched and mock/real dispatch ran.
	SMSDispatched bool
}
