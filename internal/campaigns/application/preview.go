package application

import (
	"strings"

	"skykin-platform/internal/campaigns/domain"
)

// ChannelLabel returns a human-readable label for portal preview.
func ChannelLabel(channelCode string) string {
	switch strings.ToUpper(channelCode) {
	case "SMS_PLUS":
		return "SMS+"
	case "PUSH":
		return "Push Notification"
	case "NATIVE_FEED":
		return "Native Feed"
	default:
		return "In-App Banner"
	}
}

// PreviewCampaign builds a simulator payload for the ad portal preview endpoint.
func PreviewCampaign(c *domain.Campaign, channelCode string) map[string]any {
	content, _ := CampaignAdContent(c, channelCode)
	return map[string]any{
		"format":        channelCode,
		"campaign_name": c.Name,
		"simulator":     true,
		"channel_label": ChannelLabel(channelCode),
		"preview":       content,
	}
}
