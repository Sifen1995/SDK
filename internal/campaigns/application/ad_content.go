package application

import campaigndomain "skykin-platform/internal/campaigns/domain"

// CampaignAdContent builds the SDK ad payload from a campaign.
func CampaignAdContent(c *campaigndomain.Campaign, channelCode string) (map[string]any, error) {
	canvas := c.CanvasJSON
	if canvas == nil {
		canvas = map[string]any{}
	}
	return map[string]any{
		"title":     c.Title,
		"body_text": c.BodyText,
		"image_url": c.ImageURL,
		// Preserve the advertiser's destination exactly as submitted.
		"destination_url": c.DestinationURL,
		"channel_code":    channelCode,
		"canvas_json":     canvas,
	}, nil
}
