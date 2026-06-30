package consumers

import (
	"log/slog"

	adminEvents "skykin-platform/internal/admin/events"
	"skykin-platform/internal/platform/messaging"
)

// ModerationConsumer reacts to admin campaign moderation lifecycle events.
type ModerationConsumer struct {
	log *slog.Logger
}

func NewModerationConsumer(log *slog.Logger) *ModerationConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &ModerationConsumer{log: log}
}

func (c *ModerationConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, adminEvents.TopicCampaignModerationPassed, c.handlePassed)
	messaging.Register(bus, adminEvents.TopicCampaignModerationRejected, c.handleRejected)
	messaging.Register(bus, adminEvents.TopicCampaignActivated, c.handleActivated)
}

func (c *ModerationConsumer) handlePassed(e messaging.Event) {
	evt, ok := e.Payload.(adminEvents.CampaignModerationPassedEvent)
	if !ok {
		c.log.Error("invalid campaign moderation passed payload")
		return
	}
	c.log.Info("campaign moderation passed",
		"campaign_id", evt.CampaignID,
		"operator_id", evt.OperatorUserID,
		"channel", evt.ChannelCode,
	)
}

func (c *ModerationConsumer) handleRejected(e messaging.Event) {
	evt, ok := e.Payload.(adminEvents.CampaignModerationRejectedEvent)
	if !ok {
		c.log.Error("invalid campaign moderation rejected payload")
		return
	}
	c.log.Info("campaign moderation rejected",
		"campaign_id", evt.CampaignID,
		"operator_id", evt.OperatorUserID,
	)
}

func (c *ModerationConsumer) handleActivated(e messaging.Event) {
	evt, ok := e.Payload.(adminEvents.CampaignActivatedEvent)
	if !ok {
		c.log.Error("invalid campaign activated payload")
		return
	}
	c.log.Info("campaign activated",
		"campaign_id", evt.CampaignID,
		"operator_id", evt.OperatorUserID,
	)
}
