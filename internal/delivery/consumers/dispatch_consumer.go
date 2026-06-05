package consumers

import (
	"context"

	campaignApp "skykin-platform/internal/campaigns/application"
	"skykin-platform/internal/delivery/model"
	"skykin-platform/internal/platform/messaging"
	platformWS "skykin-platform/internal/platform/websocket"
	usermodel "skykin-platform/internal/users/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const EventCampaignMatched = "CampaignMatched"

type DispatchConsumer struct {
	db         *gorm.DB
	adDelivery *campaignApp.AdDeliveryService
	hub        platformWS.Notifier
}

func NewDispatchConsumer(db *gorm.DB, adDelivery *campaignApp.AdDeliveryService, hub platformWS.Notifier) *DispatchConsumer {
	return &DispatchConsumer{db: db, adDelivery: adDelivery, hub: hub}
}

func (c *DispatchConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, EventCampaignMatched, c.handle)
}

func (c *DispatchConsumer) handle(e messaging.Event) {
	payload, ok := e.Payload.(map[string]any)
	if !ok {
		return
	}
	userID, intent, ok := parseMatchedPayload(payload)
	if !ok {
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	campaignID, _ := payload["campaign_id"].(uuid.UUID)
	_ = c.db.WithContext(ctx).Create(&model.DeliveryJob{
		UserID: userID.String(), CampaignID: campaignID.String(),
	}).Error

	var user usermodel.Users
	if err := c.db.WithContext(ctx).Where("id = ?", userID.String()).First(&user).Error; err != nil {
		return
	}
	ad, err := c.adDelivery.BuildAdForIntent(ctx, intent)
	if err != nil {
		return
	}
	_ = c.hub.NotifyUser(user.ExternalUserID, map[string]any{
		"type": ad.Type, "intent": ad.Intent, "campaign_id": ad.CampaignID,
		"campaign_name": ad.CampaignName, "creative_format": ad.CreativeFormat, "content": ad.Content,
	})
}

func parseMatchedPayload(p map[string]any) (uuid.UUID, string, bool) {
	uid, ok1 := p["user_id"].(uuid.UUID)
	intent, ok2 := p["intent"].(string)
	return uid, intent, ok1 && ok2
}
