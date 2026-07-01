package validation

import (
	"strings"

	"skykin-platform/internal/campaigns/domain"
)

// Result is stored on the campaign row as validation_status / validation_notes.
type Result struct {
	Status string
	Notes  string
}

// Campaign checks creative fields against the resolved channel code.
func Campaign(c *domain.Campaign, channelCode string) Result {
	code := strings.ToUpper(strings.TrimSpace(channelCode))
	switch code {
	case "IN_APP_BANNER", "NATIVE_FEED":
		if strings.TrimSpace(c.ImageURL) == "" {
			return Result{Status: "failed", Notes: "image_url is required for banner channels"}
		}
	case "PUSH":
		if len(c.Title) == 0 || len(c.Title) > 50 {
			return Result{Status: "failed", Notes: "title must be 1-50 characters for push"}
		}
		if len(c.BodyText) == 0 || len(c.BodyText) > 120 {
			return Result{Status: "failed", Notes: "body_text must be 1-120 characters for push"}
		}
	case "SMS_PLUS":
		if len(c.Title) == 0 || len(c.Title) > 40 {
			return Result{Status: "failed", Notes: "title must be 1-40 characters for SMS+"}
		}
		if len(c.BodyText) == 0 || len(c.BodyText) > 160 {
			return Result{Status: "failed", Notes: "body_text must be 1-160 characters for SMS+"}
		}
	default:
		return Result{Status: "failed", Notes: "unknown channel code: " + code}
	}
	return Result{Status: "passed"}
}
