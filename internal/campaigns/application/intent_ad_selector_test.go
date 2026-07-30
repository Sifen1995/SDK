package application_test

import (
	"context"
	"fmt"
	"testing"

	campaignApp "skykin-platform/internal/campaigns/application"
	campaigndomain "skykin-platform/internal/campaigns/domain"
)

type fakeCampaigns struct {
	byChannel map[string]*campaigndomain.Campaign
}

func (f *fakeCampaigns) SelectBestCampaign(
	ctx context.Context,
	intentName, channelCode, pseudonymousID string,
) (*campaigndomain.Campaign, error) {
	c, ok := f.byChannel[channelCode]
	if !ok || c == nil {
		return nil, fmt.Errorf("none")
	}
	return c, nil
}

func TestIntentAdSelector_SMSConsentedPrefersSMS(t *testing.T) {
	sel := campaignApp.NewIntentAdSelector(&fakeCampaigns{byChannel: map[string]*campaigndomain.Campaign{
		"SMS_PLUS": {
			ID: "sms", Name: "SMS", Title: "T", BodyText: "B", DestinationURL: "https://x",
			PlanMonthlyFeeETB: 10,
		},
		"IN_APP_BANNER": {
			ID: "banner", Name: "Banner", Title: "T", ImageURL: "https://img", DestinationURL: "https://x",
			PlanMonthlyFeeETB: 100,
		},
	}})

	ad, err := sel.SelectAd(context.Background(), "pseudo", "fashion_interest", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.ChannelCode != "SMS_PLUS" || ad.CampaignID != "sms" {
		t.Fatalf("expected SMS preference, got %+v", ad)
	}
}

func TestIntentAdSelector_SMSConsentedFallsBack(t *testing.T) {
	sel := campaignApp.NewIntentAdSelector(&fakeCampaigns{byChannel: map[string]*campaigndomain.Campaign{
		"IN_APP_BANNER": {
			ID: "banner", Name: "Banner", Title: "T", ImageURL: "https://img", DestinationURL: "https://x",
			PlanMonthlyFeeETB: 10,
		},
	}})

	ad, err := sel.SelectAd(context.Background(), "pseudo", "fashion_interest", "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.ChannelCode != "IN_APP_BANNER" {
		t.Fatalf("expected banner fallback, got %s", ad.ChannelCode)
	}
}

func TestIntentAdSelector_NoSMSConsentExcludesSMS(t *testing.T) {
	sel := campaignApp.NewIntentAdSelector(&fakeCampaigns{byChannel: map[string]*campaigndomain.Campaign{
		"SMS_PLUS": {
			ID: "sms", Name: "SMS", Title: "T", BodyText: "B", DestinationURL: "https://x",
			PlanMonthlyFeeETB: 999,
		},
		"IN_APP_BANNER": {
			ID: "banner", Name: "Banner", Title: "T", ImageURL: "https://img", DestinationURL: "https://x",
			PlanMonthlyFeeETB: 1,
		},
	}})

	ad, err := sel.SelectAd(context.Background(), "pseudo", "fashion_interest", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ad.ChannelCode == "SMS_PLUS" {
		t.Fatal("SMS_PLUS must be excluded when smsConsented=false")
	}
	if ad.CampaignID != "banner" {
		t.Fatalf("expected banner, got %s", ad.CampaignID)
	}
}
