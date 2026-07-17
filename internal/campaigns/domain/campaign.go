package domain

import "time"

// Campaign stores targeting, budget, creative payload, and delivery channel in one row.
type Campaign struct {
	ID           string
	AdvertiserID string

	Name         string
	TargetIntent string

	ChannelID string
	SegmentID *string

	Title          string
	BodyText       string
	ImageURL       string
	DestinationURL string
	CanvasJSON     map[string]any

	BillingModel   string
	DailyBudgetCap float64

	// Delivery ranking fields (populated when loading eligible campaigns via active subscription join).
	PlanID            string
	PlanName          string
	PlanMonthlyFeeETB float64
	ChannelCode       string // populated on delivery/master list reads via channels join
	TotalBudgetCap    float64
	BudgetSpent       float64
	FrequencyCapPerDay int

	ScheduledStartAt *time.Time
	ScheduledEndAt   *time.Time

	IsActive         bool
	ValidationStatus string
	ValidationNotes  string

	ModerationStatus string
	ModerationNotes  string
	ModeratedAt      *time.Time
	ModeratedBy      *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
