package consumers

import (
	"context"
	"strconv"

	campaignApp "skykin-platform/internal/campaigns/application"
	"skykin-platform/internal/delivery/infrastructure/persistence"
	"skykin-platform/internal/platform/messaging"
	platformWS "skykin-platform/internal/platform/websocket"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const EventCampaignMatched = "CampaignMatched"

type DispatchConsumer struct {
	db         *gorm.DB
	adDelivery *campaignApp.AdDeliveryService
	hub        platformWS.Notifier
}

func NewDispatchConsumer(db *gorm.DB, adDelivery *campaignApp.AdDeliveryService, hub *platformWS.Hub) *DispatchConsumer {
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
	_ = c.db.WithContext(ctx).Create(&persistence.DeliveryJobRow{
		UserID: userID, CampaignID: campaignID.String(),
	}).Error

	var pseudo string
	_ = c.db.WithContext(ctx).Raw(`
		SELECT pseudonymous_id::text FROM pseudonymous_mappings WHERE user_id = ?
	`, parseUserIDAsInt(userID)).Scan(&pseudo).Error
	notifyKey := pseudo
	if notifyKey == "" {
		notifyKey = userID
	}

	ad, err := c.adDelivery.BuildAdForIntent(ctx, intent)
	if err != nil {
		return
	}
	_ = c.hub.NotifyUser(notifyKey, map[string]any{
		"type": ad.Type, "intent": ad.Intent, "campaign_id": ad.CampaignID,
		"campaign_name": ad.CampaignName, "channel_code": ad.ChannelCode,
		"creative_format": ad.ChannelCode, "content": ad.Content,
	})
}

func parseMatchedPayload(p map[string]any) (string, string, bool) {
	intent, ok2 := p["intent"].(string)
	switch uid := p["user_id"].(type) {
	case string:
		return uid, intent, ok2 && uid != ""
	case uuid.UUID:
		return uid.String(), intent, ok2
	default:
		return "", intent, false
	}
}

func parseUserIDAsInt(userID string) int64 {
	n, _ := strconv.ParseInt(userID, 10, 64)
	return n
}
