package consumers

import (
	campaignEvents "skykin-platform/internal/campaigns/events"
	"skykin-platform/internal/platform/messaging"
	platformWS "skykin-platform/internal/platform/websocket"
)

type CampaignAdConsumer struct {
	hub platformWS.Notifier
}

func NewCampaignAdConsumer(hub platformWS.Notifier) *CampaignAdConsumer {
	return &CampaignAdConsumer{hub: hub}
}

func (c *CampaignAdConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, campaignEvents.TopicCampaignAdDelivered, c.handle)
}

func (c *CampaignAdConsumer) handle(e messaging.Event) {
	p, ok := e.Payload.(campaignEvents.CampaignAdDelivered)
	if !ok || p.Ad == nil {
		return
	}
	_ = c.hub.NotifyUser(p.ExternalUserID, p.Ad)
}
