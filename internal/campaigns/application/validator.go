package application

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"skykin-platform/internal/campaigns/infrastructure"
	"skykin-platform/internal/campaigns/model"
)

type ValidationResult struct {
	Status string
	Notes  string
}

func NormalizeCreativeFormat(raw string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "BANNER", "BANNER_IMAGE":
		return "BANNER", nil
	case "PUSH", "PUSH_PLUS", "PUSH_NOTIFICATION":
		return "PUSH_PLUS", nil
	case "SMS", "SMS_PLUS", "SMS+":
		return "SMS_PLUS", nil
	default:
		return "", fmt.Errorf("creative_format must be BANNER, PUSH_PLUS, or SMS_PLUS")
	}
}

func ValidateCampaign(c *model.Campaign) ValidationResult {
	if strings.TrimSpace(c.DestinationURL) == "" {
		return ValidationResult{Status: "failed", Notes: "destination_url is required"}
	}
	if !reachableURL(c.DestinationURL) {
		return ValidationResult{Status: "warning", Notes: "destination_url is not reachable"}
	}

	switch c.CreativeFormat {
	case "BANNER":
		if strings.TrimSpace(c.ImageURL) == "" {
			return ValidationResult{Status: "failed", Notes: "image_url is required for BANNER"}
		}
	case "PUSH_PLUS":
		if len(c.Title) == 0 || len(c.Title) > 50 {
			return ValidationResult{Status: "failed", Notes: "title must be 1-50 characters for PUSH_PLUS"}
		}
		if len(c.BodyText) == 0 || len(c.BodyText) > 120 {
			return ValidationResult{Status: "failed", Notes: "body_text must be 1-120 characters for PUSH_PLUS"}
		}
	case "SMS_PLUS":
		if len(c.Title) == 0 || len(c.Title) > 40 {
			return ValidationResult{Status: "failed", Notes: "title must be 1-40 characters for SMS_PLUS"}
		}
		if len(c.BodyText) == 0 || len(c.BodyText) > 160 {
			return ValidationResult{Status: "failed", Notes: "body_text must be 1-160 characters for SMS_PLUS"}
		}
	}

	return ValidationResult{Status: "passed"}
}

func reachableURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodHead, raw, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		resp, err = client.Get(raw)
		if err != nil {
			return false
		}
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func ChannelLabel(format string) string {
	switch format {
	case "SMS_PLUS":
		return "SMS+"
	case "PUSH_PLUS":
		return "Push Notification"
	default:
		return "In-App Banner"
	}
}

func PreviewCampaign(c *model.Campaign) map[string]any {
	content, _ := infrastructure.CampaignAdContent(c)
	return map[string]any{
		"format":        c.CreativeFormat,
		"campaign_name": c.Name,
		"simulator":     true,
		"channel_label": ChannelLabel(c.CreativeFormat),
		"preview":       content,
	}
}
