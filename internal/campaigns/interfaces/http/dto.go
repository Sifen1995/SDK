package http

import "time"

// CreateCampaignRequest is the HTTP transport DTO (Gin binding). Mapped to application.CreateCampaignCommand in the handler.
type CreateCampaignRequest struct {
	Name         string `json:"name" binding:"required,min=3,max=255"`
	TargetIntent string `json:"target_intent" binding:"required"`
	ChannelID    string `json:"channel_id" binding:"required,uuid"`

	// SegmentID is optional (nil means target by intent only via the free tier)
	SegmentID *string `json:"segment_id" binding:"omitempty,uuid"`

	// Creative Asset Payloads
	Title          string `json:"title" binding:"max=255"`
	BodyText       string `json:"body_text" binding:"required"`
	ImageURL       string `json:"image_url" binding:"omitempty,url"`
	DestinationURL string `json:"destination_url" binding:"required,url"`

	// CanvasJSON takes raw interactive layout maps from the SMS+ builder UI
	CanvasJSON map[string]interface{} `json:"canvas_json" binding:"omitempty"`

	// Financial & Operational Controls. Billing is standardized to CPC and is
	// intentionally not accepted from the portal.
	DailyBudgetCap     float64 `json:"daily_budget_cap" binding:"required,numeric,gt=0"`
	TotalBudgetCap     float64 `json:"total_budget_cap" binding:"required,numeric,gt=0,gtefield=DailyBudgetCap"`
	FrequencyCapPerDay int     `json:"frequency_cap_per_day" binding:"required,numeric,min=1"`

	// Scheduling Options (RFC3339 formatted strings)
	ScheduledStartAt *time.Time `json:"scheduled_start_at" binding:"omitempty"`
	ScheduledEndAt   *time.Time `json:"scheduled_end_at" binding:"omitempty,gtfield=ScheduledStartAt"`
}
