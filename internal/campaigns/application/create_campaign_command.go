package application

import "time"

// CreateCampaignCommand is the application-layer input for campaign creation.
// HTTP handlers map CreateCampaignRequest → this struct (no Gin/JSON tags here).
type CreateCampaignCommand struct {
	Name               string
	TargetIntent       string
	ChannelID          string
	SegmentID          *string // nil = free intent-only targeting (no Audiencemart purchase)
	Title              string
	BodyText           string
	ImageURL           string
	DestinationURL     string
	CanvasJSON         map[string]any
	BillingModel       string
	DailyBudgetCap     float64
	TotalBudgetCap     float64
	FrequencyCapPerDay int
	ScheduledStartAt   *time.Time
	ScheduledEndAt     *time.Time
}
