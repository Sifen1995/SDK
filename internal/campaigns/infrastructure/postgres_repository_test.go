package infrastructure

import (
	"testing"

	"skykin-platform/internal/campaigns/domain"
)

func TestCampaignAdContentPreservesDestinationURL(t *testing.T) {
	destination := "https://merchant.example/promo?source=summer"
	content, err := CampaignAdContent(&domain.Campaign{
		ID:             "550e8400-e29b-41d4-a716-446655440000",
		DestinationURL: destination,
	}, "IN_APP_BANNER", NewPlayLinkBuilder("click-secret"))
	if err != nil {
		t.Fatalf("build campaign content: %v", err)
	}
	if got := content["destination_url"]; got != destination {
		t.Fatalf("destination_url = %q, want unchanged %q", got, destination)
	}
}
