package http

type CreateCampaignRequest struct {
	Name           string                 `json:"name" binding:"required" example:"Crypto Promo"`
	TargetIntent   string                 `json:"target_intent" binding:"required" example:"crypto_interest"`
	ApplicationID  string                 `json:"application_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreativeFormat string                 `json:"creative_format" binding:"required" example:"BANNER"`
	Title          string                 `json:"title" example:"Save on fees"`
	BodyText       string                 `json:"body_text" example:"Trade crypto with zero fees today"`
	ImageURL       string                 `json:"image_url" example:"https://cdn.example.com/banner.png"`
	DestinationURL string                 `json:"destination_url" binding:"required" example:"https://example.com/offer"`
	CanvasJSON     map[string]interface{} `json:"canvas_json"`
	DailyBudgetCap float64                `json:"daily_budget_cap" example:"100"`
	TotalBudgetCap float64                `json:"total_budget_cap" example:"1000"`
}
