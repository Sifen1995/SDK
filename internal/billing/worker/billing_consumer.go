package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	billingdomain "skykin-platform/internal/billing/domain"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	campaigndomain "skykin-platform/internal/campaigns/domain"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

const (
	billingEventsStream      = "stream:billing_events"
	billingProcessorGroup    = "billing_processor_group"
	billingProcessorConsumer = "billing-worker-1"
	billingReadBatchSize     = 100
	billingReadBlock         = 2 * time.Second
)

// StartBillingConsumer creates the Redis consumer group and processes stream:billing_events.
// This worker only persists billing_events (+ Redis budget keys). Delivery logs are owned by
// the delivery module's consumer group on the same stream.
func StartBillingConsumer(db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	if db == nil || rdb == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	if err := rdb.XGroupCreateMkStream(context.Background(), billingEventsStream, billingProcessorGroup, "0"); err != nil {
		logger.Error("billing consumer: create group failed", "error", err)
		return
	}
	logger.Info("billing consumer: group ready",
		"stream", billingEventsStream,
		"group", billingProcessorGroup,
	)
	go runBillingConsumer(context.Background(), db, rdb, logger)
}

func runBillingConsumer(ctx context.Context, db *gorm.DB, rdb *platformredis.RedisClient, logger *slog.Logger) {
	campaigns := campaignInfra.NewRepository(db)
	subs := billingInfra.NewSubscriptionRepository(db)
	rates := billingInfra.NewBillingRateRepository(db)
	budgetFlags := campaignInfra.NewRedisCampaignRepository(rdb)

	for {
		if ctx.Err() != nil {
			return
		}

		msgs, err := rdb.XReadGroup(
			ctx,
			billingProcessorGroup,
			billingProcessorConsumer,
			billingEventsStream,
			"0",
			billingReadBatchSize,
			0,
		)
		if err != nil && err != platformredis.ErrNil {
			logger.Warn("billing consumer: xreadgroup pending failed", "error", err)
			time.Sleep(time.Second)
			continue
		}
		if len(msgs) == 0 {
			msgs, err = rdb.XReadGroup(
				ctx,
				billingProcessorGroup,
				billingProcessorConsumer,
				billingEventsStream,
				">",
				billingReadBatchSize,
				billingReadBlock,
			)
			if err != nil {
				if err == platformredis.ErrNil {
					continue
				}
				logger.Warn("billing consumer: xreadgroup failed", "error", err)
				time.Sleep(time.Second)
				continue
			}
		}
		if len(msgs) == 0 {
			continue
		}

		ids, err := processBillingBatch(ctx, db, rdb, campaigns, subs, rates, budgetFlags, msgs, logger)
		if err != nil {
			logger.Error("billing consumer: batch failed", "count", len(msgs), "error", err)
			continue
		}
		if err := rdb.XAck(ctx, billingEventsStream, billingProcessorGroup, ids...); err != nil {
			logger.Error("billing consumer: xack failed", "ids", len(ids), "error", err)
			continue
		}
		logger.Info("billing consumer: processed batch", "count", len(ids))
	}
}

func processBillingBatch(
	ctx context.Context,
	db *gorm.DB,
	rdb *platformredis.RedisClient,
	campaigns *campaignInfra.Repository,
	subs *billingInfra.SubscriptionRepository,
	rates *billingInfra.BillingRateRepository,
	budgetFlags *campaignInfra.RedisCampaignRepository,
	msgs []platformredis.StreamMessage,
	logger *slog.Logger,
) ([]string, error) {
	campaignCache := map[string]*campaigndomain.Campaign{}
	subCache := map[string]*billingdomain.AdvertiserSubscription{}
	rateCache := map[string][]billingdomain.BillingRate{}

	built := make([]billingdomain.BillingEvent, 0, len(msgs))
	ids := make([]string, 0, len(msgs))
	chargesByCampaign := map[string]float64{}
	skipped := 0

	for _, msg := range msgs {
		evt, campaign, err := resolveStreamMessage(ctx, campaigns, subs, rates, campaignCache, subCache, rateCache, msg)
		if err != nil {
			skipped++
			logger.Warn("billing consumer: skip message",
				"id", msg.ID,
				"campaign_id", strings.TrimSpace(msg.Values["campaign_id"]),
				"event_type", strings.TrimSpace(msg.Values["event_type"]),
				"error", err,
			)
			ids = append(ids, msg.ID)
			continue
		}
		built = append(built, *evt)
		ids = append(ids, msg.ID)
		chargesByCampaign[campaign.ID] += evt.ChargeETB
	}

	if len(built) > 0 {
		if err := billingInfra.NewBillingEventRepository(db).CreateBatch(ctx, built); err != nil {
			return nil, err
		}
	}

	if skipped > 0 {
		logger.Warn("billing consumer: skipped invalid stream messages",
			"skipped", skipped,
			"accepted", len(built),
		)
	}

	today := time.Now().UTC().Format("2006-01-02")
	for campaignID, charge := range chargesByCampaign {
		if charge <= 0 {
			continue
		}
		key := fmt.Sprintf("budget:spent:%s:%s", campaignID, today)
		spent, err := rdb.IncrByFloat(ctx, key, charge)
		if err != nil {
			logger.Warn("billing consumer: budget incr failed", "campaign_id", campaignID, "error", err)
			continue
		}
		_ = rdb.Expire(ctx, key, 48*time.Hour)

		campaign := campaignCache[campaignID]
		if campaign == nil {
			continue
		}
		if budgetExceeded(campaign, spent) {
			if err := budgetFlags.SetBudgetExhausted(ctx, campaignID, 0); err != nil {
				logger.Warn("billing consumer: set budget_exhausted failed", "campaign_id", campaignID, "error", err)
			} else {
				logger.Info("billing consumer: campaign budget exhausted",
					"campaign_id", campaignID,
					"spent", spent,
					"daily_cap", campaign.DailyBudgetCap,
					"total_cap", campaign.TotalBudgetCap,
				)
			}
		}
	}

	return ids, nil
}

func resolveStreamMessage(
	ctx context.Context,
	campaigns *campaignInfra.Repository,
	subs *billingInfra.SubscriptionRepository,
	rates *billingInfra.BillingRateRepository,
	campaignCache map[string]*campaigndomain.Campaign,
	subCache map[string]*billingdomain.AdvertiserSubscription,
	rateCache map[string][]billingdomain.BillingRate,
	msg platformredis.StreamMessage,
) (*billingdomain.BillingEvent, *campaigndomain.Campaign, error) {
	campaignID := strings.TrimSpace(msg.Values["campaign_id"])
	eventType := strings.ToLower(strings.TrimSpace(msg.Values["event_type"]))
	if campaignID == "" || eventType == "" {
		return nil, nil, fmt.Errorf("campaign_id and event_type are required")
	}

	campaign, ok := campaignCache[campaignID]
	if !ok {
		c, err := campaigns.Get(ctx, campaignID)
		if err != nil {
			return nil, nil, fmt.Errorf("load campaign: %w", err)
		}
		campaign = c
		campaignCache[campaignID] = c
	}

	sub, ok := subCache[campaign.AdvertiserID]
	if !ok {
		s, err := subs.GetActiveByAdvertiser(ctx, campaign.AdvertiserID)
		if err != nil {
			return nil, nil, fmt.Errorf("load subscription: %w", err)
		}
		sub = s
		subCache[campaign.AdvertiserID] = s
	}

	planRates, ok := rateCache[sub.PlanID]
	if !ok {
		list, err := rates.ListByPlanID(ctx, sub.PlanID)
		if err != nil {
			return nil, nil, fmt.Errorf("load rates: %w", err)
		}
		planRates = list
		rateCache[sub.PlanID] = list
	}

	// Campaigns no longer store billing model. The legacy telemetry contract
	// also does not include it, so resolve the model from the event type.
	model := defaultModelForEvent(eventType)
	rate, ok := findRate(planRates, eventType, model)
	if !ok {
		return nil, nil, fmt.Errorf("no billing rate for plan=%s event=%s model=%s", sub.PlanID, eventType, model)
	}

	txn, _ := strconv.ParseFloat(strings.TrimSpace(msg.Values["transaction_value"]), 64)
	occurredAt := parseOccurredAt(msg.Values["occurred_at"])
	charge := computeCharge(model, rate.RateETB, txn)

	return &billingdomain.BillingEvent{
		AdvertiserID:     campaign.AdvertiserID,
		CampaignID:       campaign.ID,
		SubscriptionID:   sub.ID,
		EventType:        eventType,
		BillingModel:     model,
		RateApplied:      rate.RateETB,
		TransactionValue: txn,
		ChargeETB:        charge,
		IsBilled:         false,
		OccurredAt:       occurredAt,
	}, campaign, nil
}

func findRate(rates []billingdomain.BillingRate, eventType, model string) (billingdomain.BillingRate, bool) {
	for _, r := range rates {
		if !r.IsActive {
			continue
		}
		if strings.EqualFold(r.EventType, eventType) && strings.EqualFold(r.Model, model) {
			return r, true
		}
	}
	return billingdomain.BillingRate{}, false
}

func defaultModelForEvent(eventType string) string {
	switch eventType {
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

func computeCharge(model string, rateETB, transactionValue float64) float64 {
	switch strings.ToUpper(model) {
	case "CPM":
		return rateETB / 1000.0
	case "REV_SHARE":
		return transactionValue * (rateETB / 100.0)
	default:
		return rateETB
	}
}

func parseOccurredAt(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC()
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC()
	}
	return time.Now().UTC()
}

func budgetExceeded(c *campaigndomain.Campaign, daySpent float64) bool {
	if c == nil {
		return false
	}
	if c.DailyBudgetCap > 0 && daySpent >= c.DailyBudgetCap {
		return true
	}
	totalApprox := c.BudgetSpent + daySpent
	if c.TotalBudgetCap > 0 && totalApprox >= c.TotalBudgetCap {
		return true
	}
	return false
}
