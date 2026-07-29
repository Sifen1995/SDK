package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	billingdomain "skykin-platform/internal/billing/domain"
	billingInfra "skykin-platform/internal/billing/infrastructure"
	campaignInfra "skykin-platform/internal/campaigns/infrastructure"
	deliverydomain "skykin-platform/internal/delivery/domain"
	deliveryInfra "skykin-platform/internal/delivery/infrastructure"
	"skykin-platform/internal/platform/redis"
)

type CPCWorker struct {
	rdb    *redis.RedisClient
	db     *gorm.DB
	logger *slog.Logger
}

type ClickQueuePayload struct {
	CampaignID   string    `json:"campaign_id"`
	ClickedAt    time.Time `json:"clicked_at"`
	EventType    string    `json:"event_type"`
	BillingModel string    `json:"billing_model"`
}

func NewCPCWorker(rdb *redis.RedisClient, db *gorm.DB, logger *slog.Logger) *CPCWorker {
	return &CPCWorker{rdb: rdb, db: db, logger: logger}
}

func (w *CPCWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			res, err := w.rdb.BRPop(ctx, "queue:cpc_billing_events", 2*time.Second)
			if err != nil || res == "" {
				continue
			}

			var payload ClickQueuePayload
			if err := json.Unmarshal([]byte(res), &payload); err != nil {
				w.logger.Error("failed to parse CPC click payload", "error", err)
				continue
			}

			if err := w.processCPCClick(ctx, payload); err != nil {
				w.logger.Error("failed to log CPC billing entry", "error", err)
			}
		}
	}
}

func (w *CPCWorker) processCPCClick(ctx context.Context, payload ClickQueuePayload) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Fetch campaign to get advertiser_id
		campaignRepo := campaignInfra.NewRepository(tx)
		campaign, err := campaignRepo.Get(ctx, payload.CampaignID)
		if err != nil {
			return fmt.Errorf("fetch campaign: %w", err)
		}
		if campaign == nil {
			return fmt.Errorf("campaign not found: %s", payload.CampaignID)
		}

		// 2. Fetch subscription for the advertiser
		subRepo := billingInfra.NewSubscriptionRepository(tx)
		sub, err := subRepo.GetActiveByAdvertiser(ctx, campaign.AdvertiserID)
		if err != nil {
			return fmt.Errorf("fetch subscription: %w", err)
		}
		if sub == nil {
			return fmt.Errorf("no active subscription for advertiser: %s", campaign.AdvertiserID)
		}

		// 3. Fetch the rate for the model supplied by the tracking endpoint.
		rateRepo := billingInfra.NewBillingRateRepository(tx)
		rates, err := rateRepo.ListByPlanID(ctx, sub.PlanID)
		if err != nil {
			return fmt.Errorf("fetch billing rates: %w", err)
		}

		model := strings.ToUpper(strings.TrimSpace(payload.BillingModel))
		if model == "" {
			return fmt.Errorf("billing_model is required")
		}
		var billingRate billingdomain.BillingRate
		found := false
		for _, r := range rates {
			if r.IsActive && strings.EqualFold(r.EventType, "click") && strings.EqualFold(r.Model, model) {
				billingRate = r
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("no active %s click rate found for plan: %s", model, sub.PlanID)
		}

		// 4. Create and persist DeliveryLog
		deliveryLog := deliverydomain.DeliveryLog{
			ID:             uuid.New().String(),
			CampaignID:     payload.CampaignID,
			UserID:         deliverydomain.AnonymousUserID,
			SessionID:      "ANONYMOUS_CPC",
			DeliveryStatus: deliverydomain.StatusClicked,
			LoggedAt:       payload.ClickedAt,
		}
		deliveryLogRepo := deliveryInfra.NewDeliveryLogRepository(tx)
		if err := deliveryLogRepo.CreateBatch(ctx, []deliverydomain.DeliveryLog{deliveryLog}); err != nil {
			return fmt.Errorf("persist delivery log: %w", err)
		}

		// 5. Create and persist BillingEvent
		billingEvent := billingdomain.BillingEvent{
			ID:               uuid.New().String(),
			AdvertiserID:     campaign.AdvertiserID,
			CampaignID:       payload.CampaignID,
			SubscriptionID:   sub.ID,
			EventType:        "click",
			BillingModel:     model,
			RateApplied:      billingRate.RateETB,
			TransactionValue: 0,
			ChargeETB:        clickCharge(model, billingRate.RateETB),
			IsBilled:         false,
			OccurredAt:       payload.ClickedAt,
			CreatedAt:        time.Now().UTC(),
		}
		billingEventRepo := billingInfra.NewBillingEventRepository(tx)
		if err := billingEventRepo.CreateBatch(ctx, []billingdomain.BillingEvent{billingEvent}); err != nil {
			return fmt.Errorf("persist billing event: %w", err)
		}

		return nil
	})
}

func clickCharge(model string, rateETB float64) float64 {
	if strings.EqualFold(model, "CPM") {
		return rateETB / 1000
	}
	return rateETB
}
