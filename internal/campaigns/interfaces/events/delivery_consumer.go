package events

import (
	"context"
	"log/slog"

	campaignApp "skykin-platform/internal/campaigns/application"
	campaignEvents "skykin-platform/internal/campaigns/events"
	"skykin-platform/internal/platform/messaging"
)

// DeliveryConsumer logs campaign ad dispatches inside the campaigns bounded context.
type DeliveryConsumer struct {
	delivery *campaignApp.AdDeliveryService
	log      *slog.Logger
}

func NewDeliveryConsumer(delivery *campaignApp.AdDeliveryService, log *slog.Logger) *DeliveryConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &DeliveryConsumer{delivery: delivery, log: log}
}

func (c *DeliveryConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, campaignEvents.TopicCampaignAdDelivered, c.handle)
}

func (c *DeliveryConsumer) handle(e messaging.Event) {
	evt, ok := e.Payload.(campaignEvents.CampaignAdDelivered)
	if !ok {
		c.log.Error("invalid campaign ad delivered payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	campaignID := ""
	if evt.Ad != nil {
		if id, ok := evt.Ad["campaign_id"].(string); ok {
			campaignID = id
		}
	}
	if campaignID == "" || evt.InternalUserID == "" {
		return
	}
	sessionID := evt.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}
	if err := c.delivery.LogDispatched(ctx, campaignID, evt.InternalUserID, sessionID); err != nil {
		c.log.Error("log delivery failed", "campaign_id", campaignID, "error", err)
	}
}
