package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
)

// CampaignBillingInfo is the narrow campaign projection billing needs.
type CampaignBillingInfo struct {
	ID               string
	AdvertiserID     string
	DailyBudgetCap   float64
	TotalBudgetCap   float64
	CurrentBudgetUse float64
}

// CampaignBillingReader prevents billing from importing campaign repositories.
type CampaignBillingReader interface {
	GetCampaignBillingInfo(ctx context.Context, campaignID string) (*CampaignBillingInfo, error)
}

// DailySpendTracker owns the Redis counter used by billing.
type DailySpendTracker interface {
	Add(ctx context.Context, campaignID string, amount float64, ttl time.Duration) (float64, error)
}

// BudgetExhaustionMarker tells campaign delivery to stop serving an exhausted campaign.
type BudgetExhaustionMarker interface {
	MarkExhausted(ctx context.Context, campaignID string) error
}

// BillingInput is the transport-neutral input consumed from telemetry.
type BillingInput struct {
	CampaignID      string
	EventType       string
	TransactionRaw  string
	OccurredAtRaw   string
}

// EventProcessor owns billing model selection, rate lookup, charge calculation,
// persistence, and budget exhaustion decisions.
type EventProcessor struct {
	campaigns CampaignBillingReader
	subs      billingdomain.SubscriptionRepository
	rates     billingdomain.BillingRateRepository
	events    billingdomain.BillingEventRepository
	spend     DailySpendTracker
	marker    BudgetExhaustionMarker
}

func NewEventProcessor(
	campaigns CampaignBillingReader,
	subs billingdomain.SubscriptionRepository,
	rates billingdomain.BillingRateRepository,
	events billingdomain.BillingEventRepository,
	spend DailySpendTracker,
	marker BudgetExhaustionMarker,
) *EventProcessor {
	return &EventProcessor{
		campaigns: campaigns,
		subs:      subs,
		rates:     rates,
		events:    events,
		spend:     spend,
		marker:    marker,
	}
}

func (p *EventProcessor) Process(ctx context.Context, in BillingInput) error {
	campaignID := strings.TrimSpace(in.CampaignID)
	eventType := strings.ToLower(strings.TrimSpace(in.EventType))
	if campaignID == "" || eventType == "" {
		return fmt.Errorf("campaign_id and event_type are required")
	}

	campaign, err := p.campaigns.GetCampaignBillingInfo(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("load campaign billing info: %w", err)
	}
	sub, err := p.subs.GetActiveByAdvertiser(ctx, campaign.AdvertiserID)
	if err != nil {
		return fmt.Errorf("load subscription: %w", err)
	}
	planRates, err := p.rates.ListByPlanID(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("load rates: %w", err)
	}

	model := BillingModelForEvent(eventType)
	rate, ok := ActiveRate(planRates, eventType, model)
	if !ok {
		return fmt.Errorf("no billing rate for plan=%s event=%s model=%s", sub.PlanID, eventType, model)
	}

	transactionValue, _ := strconv.ParseFloat(strings.TrimSpace(in.TransactionRaw), 64)
	event := billingdomain.BillingEvent{
		AdvertiserID:     campaign.AdvertiserID,
		CampaignID:       campaign.ID,
		SubscriptionID:   sub.ID,
		EventType:        eventType,
		BillingModel:     model,
		RateApplied:      rate.RateETB,
		TransactionValue: transactionValue,
		ChargeETB:        CalculateCharge(model, rate.RateETB, transactionValue),
		IsBilled:         false,
		OccurredAt:       ParseOccurredAt(in.OccurredAtRaw),
	}
	if err := p.events.CreateBatch(ctx, []billingdomain.BillingEvent{event}); err != nil {
		return fmt.Errorf("persist billing event: %w", err)
	}

	if event.ChargeETB <= 0 || p.spend == nil {
		return nil
	}
	spent, err := p.spend.Add(ctx, campaign.ID, event.ChargeETB, 48*time.Hour)
	if err != nil {
		return fmt.Errorf("record daily spend: %w", err)
	}
	if BudgetExceeded(campaign, spent) && p.marker != nil {
		if err := p.marker.MarkExhausted(ctx, campaign.ID); err != nil {
			return fmt.Errorf("mark campaign budget exhausted: %w", err)
		}
	}
	return nil
}

func BillingModelForEvent(eventType string) string {
	switch strings.ToLower(strings.TrimSpace(eventType)) {
	case "impression":
		return "CPM"
	case "click":
		return "CPC"
	case "install":
		return "CPI"
	case "signup":
		return "CPA"
	case "purchase":
		return "REV_SHARE"
	default:
		return "CPC"
	}
}

func ActiveRate(rates []billingdomain.BillingRate, eventType, model string) (billingdomain.BillingRate, bool) {
	for _, rate := range rates {
		if rate.IsActive &&
			strings.EqualFold(rate.EventType, eventType) &&
			strings.EqualFold(rate.Model, model) {
			return rate, true
		}
	}
	return billingdomain.BillingRate{}, false
}

func CalculateCharge(model string, rateETB, transactionValue float64) float64 {
	switch strings.ToUpper(strings.TrimSpace(model)) {
	case "CPM":
		return rateETB / 1000
	case "REV_SHARE":
		return transactionValue * rateETB / 100
	default:
		return rateETB
	}
}

func BudgetExceeded(campaign *CampaignBillingInfo, dailySpend float64) bool {
	if campaign == nil {
		return false
	}
	if campaign.DailyBudgetCap > 0 && dailySpend >= campaign.DailyBudgetCap {
		return true
	}
	return campaign.TotalBudgetCap > 0 &&
		campaign.CurrentBudgetUse+dailySpend >= campaign.TotalBudgetCap
}

func ParseOccurredAt(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC()
		}
	}
	return time.Now().UTC()
}
