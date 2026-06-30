package events

import (
	"context"
	"log/slog"

	adminEvents "skykin-platform/internal/admin/events"
	billingApp "skykin-platform/internal/billing/application"
	billingdomain "skykin-platform/internal/billing/domain"
	"skykin-platform/internal/platform/messaging"
)

// PlanConsumer seeds default billing rates when admin creates a subscription plan.
type PlanConsumer struct {
	rates billingdomain.BillingRateRepository
	log   *slog.Logger
}

func NewPlanConsumer(rates billingdomain.BillingRateRepository, log *slog.Logger) *PlanConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &PlanConsumer{rates: rates, log: log}
}

func (c *PlanConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, adminEvents.TopicSubscriptionPlanCreated, c.handle)
}

func (c *PlanConsumer) handle(e messaging.Event) {
	evt, ok := e.Payload.(adminEvents.SubscriptionPlanCreatedEvent)
	if !ok {
		c.log.Error("invalid subscription plan created payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := billingApp.SeedDefaultPlanRates(ctx, c.rates, evt.PlanID); err != nil {
		c.log.Error("seed default plan rates failed", "plan_id", evt.PlanID, "error", err)
		return
	}
	c.log.Info("default billing rates seeded", "plan_id", evt.PlanID)
}
