package consumers

import (
	"fmt"
	"log/slog"
	"strings"

	deliveryapp "skykin-platform/internal/delivery/application"
	"skykin-platform/internal/platform/messaging"
)

const eventCampaignMatched = "CampaignMatched"

type SMSPlusConsumer struct {
	dispatch *deliveryapp.SMSDispatchService
	logger   *slog.Logger
}

func NewSMSPlusConsumer(
	dispatch *deliveryapp.SMSDispatchService,
	logger *slog.Logger,
) *SMSPlusConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &SMSPlusConsumer{dispatch: dispatch, logger: logger}
}

func (c *SMSPlusConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, eventCampaignMatched, c.handle)
}

func (c *SMSPlusConsumer) handle(e messaging.Event) {
	payload, ok := e.Payload.(map[string]any)
	if !ok {
		c.logger.Error("invalid CampaignMatched payload")
		return
	}
	if strings.TrimSpace(fmt.Sprint(payload["channel"])) != "sms_plus" {
		return
	}
	campaignID := strings.TrimSpace(fmt.Sprint(payload["campaign_id"]))
	pseudonymousID := strings.TrimSpace(fmt.Sprint(payload["pseudonymous_id"]))
	if err := c.dispatch.DispatchCampaignMatch(e.Ctx, campaignID, pseudonymousID); err != nil {
		c.logger.Error("sms plus dispatch failed",
			"campaign_id", campaignID,
			"pseudonymous_id", pseudonymousID,
			"error", err,
		)
	}
}
