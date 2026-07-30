package http

// AnonymousCampaignDTO is the SDK payload for one active campaign on the anonymous path.
type AnonymousCampaignDTO struct {
	ID                 string         `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name               string         `json:"name" example:"Summer Fashion Drop"`
	TargetIntent       string         `json:"target_intent" example:"fashion_interest"`
	ChannelCode        string         `json:"channel_code" example:"IN_APP_BANNER"`
	Title              string         `json:"title" example:"New season styles"`
	BodyText           string         `json:"body_text" example:"Shop the drop"`
	ImageURL           string         `json:"image_url" example:"https://cdn.example.com/ad.png"`
	DestinationURL     string         `json:"destination_url" example:"https://shop.example.com"`
	CanvasJSON         map[string]any `json:"canvas_json"`
	FrequencyCapPerDay int            `json:"frequency_cap_per_day" example:"3"`
	ClickToken         string         `json:"click_token" example:"a1b2c3d4....2026-07-30-08"`
	PlanName           string         `json:"plan_name,omitempty" example:"Growth"`
	PlanMonthlyFeeETB  float64        `json:"plan_monthly_fee_etb,omitempty" example:"15000"`
}
