package application

import (
	"strings"

	"skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/campaigns/model"
)

// ValidationResult is stored on the campaign row as validation_status / validation_notes.
type ValidationResult struct {
	Status string
	Notes  string
}

// ValidateCampaign checks creative fields against the resolved channel code.
func ValidateCampaign(c *model.Campaign, channelCode string) ValidationResult {
	code := strings.ToUpper(strings.TrimSpace(channelCode))
	switch code {
	case "IN_APP_BANNER", "NATIVE_FEED":
		if strings.TrimSpace(c.ImageURL) == "" {
			return ValidationResult{Status: "failed", Notes: "image_url is required for banner channels"}
		}
	case "PUSH":
		if len(c.Title) == 0 || len(c.Title) > 50 {
			return ValidationResult{Status: "failed", Notes: "title must be 1-50 characters for push"}
		}
		if len(c.BodyText) == 0 || len(c.BodyText) > 120 {
			return ValidationResult{Status: "failed", Notes: "body_text must be 1-120 characters for push"}
		}
	case "SMS_PLUS":
		if len(c.Title) == 0 || len(c.Title) > 40 {
			return ValidationResult{Status: "failed", Notes: "title must be 1-40 characters for SMS+"}
		}
		if len(c.BodyText) == 0 || len(c.BodyText) > 160 {
			return ValidationResult{Status: "failed", Notes: "body_text must be 1-160 characters for SMS+"}
		}
		if c.CanvasJSON == nil || string(c.CanvasJSON) == "{}" || string(c.CanvasJSON) == "null" {
			// Canvas optional for SMS+ when body/title present; only warn if completely empty creative
		}
	default:
		return ValidationResult{Status: "failed", Notes: "unknown channel code: " + code}
	}
	return ValidationResult{Status: "passed"}
}

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
func PreviewCampaign(c *model.Campaign, channelCode string) map[string]any {
	content, _ := infrastructure.CampaignAdContent(c, channelCode)
	return map[string]any{
		"format":        channelCode,
		"campaign_name": c.Name,
		"simulator":     true,
		"channel_label": ChannelLabel(channelCode),
		"preview":       content,
	}
}
