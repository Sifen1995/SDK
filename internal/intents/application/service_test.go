package application_test

import (
	"context"
	"errors"
	"testing"

	campaigndomain "skykin-platform/internal/campaigns/domain"
	"skykin-platform/internal/intents/application"
	"skykin-platform/internal/intents/domain"
)

type stubProfiles struct{}

func (stubProfiles) Save(ctx context.Context, profile *domain.IntentProfile) error { return nil }

type stubAds struct {
	lastSMS bool
	ad      *application.AdSelection
	err     error
}

func (s *stubAds) SelectAd(
	ctx context.Context,
	pseudonymousID, targetIntent, channelCode string,
	smsConsented bool,
) (*application.AdSelection, error) {
	s.lastSMS = smsConsented
	if s.err != nil {
		return nil, s.err
	}
	return s.ad, nil
}

type stubSMS struct {
	called       bool
	campaignID   string
	pseudonymous string
	err          error
}

func (s *stubSMS) Dispatch(ctx context.Context, campaign *campaigndomain.Campaign, pseudonymousID string) error {
	s.called = true
	s.pseudonymous = pseudonymousID
	if campaign != nil {
		s.campaignID = campaign.ID
	}
	return s.err
}

func TestIngestAndFetchAd_SMSFound_Dispatches(t *testing.T) {
	ads := &stubAds{ad: &application.AdSelection{
		CampaignID: "camp-sms", CampaignName: "SMS", ChannelCode: "SMS_PLUS",
		Content:  map[string]any{"body": "hi"},
		Campaign: &campaigndomain.Campaign{ID: "camp-sms", Title: "T", BodyText: "B"},
	}}
	sms := &stubSMS{}
	svc := application.NewIntentService(stubProfiles{}, nil, ads, sms)

	result, err := svc.IngestAndFetchAd(context.Background(), &domain.IntentProfile{
		PseudonymousID: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		IntentName:     "fashion_interest",
		Confidence:     0.9,
	}, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ads.lastSMS {
		t.Fatal("expected smsConsented passed to selector")
	}
	if !sms.called || sms.campaignID != "camp-sms" {
		t.Fatalf("expected SMS dispatch for camp-sms, got called=%v id=%s", sms.called, sms.campaignID)
	}
	if !result.SMSDispatched {
		t.Fatal("expected SMSDispatched")
	}
	if result.AdContent != nil {
		t.Fatal("expected no in-app ad content after SMS dispatch")
	}
}

func TestIngestAndFetchAd_SMSMissing_ReturnsNonSMS(t *testing.T) {
	ads := &stubAds{ad: &application.AdSelection{
		CampaignID: "camp-banner", CampaignName: "Banner", ChannelCode: "IN_APP_BANNER",
		Content:  map[string]any{"title": "Sale"},
		Campaign: &campaigndomain.Campaign{ID: "camp-banner"},
	}}
	sms := &stubSMS{}
	svc := application.NewIntentService(stubProfiles{}, nil, ads, sms)

	result, err := svc.IngestAndFetchAd(context.Background(), &domain.IntentProfile{
		PseudonymousID: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		IntentName:     "fashion_interest",
		Confidence:     0.9,
	}, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sms.called {
		t.Fatal("dispatcher must not be called for non-SMS ad")
	}
	if result.SMSDispatched {
		t.Fatal("SMSDispatched should be false")
	}
	if result.ChannelCode != "IN_APP_BANNER" {
		t.Fatalf("got channel %s", result.ChannelCode)
	}
}

func TestIngestAndFetchAd_SMSConsentedFalse_NeverDispatches(t *testing.T) {
	ads := &stubAds{ad: &application.AdSelection{
		CampaignID: "camp-banner", CampaignName: "Banner", ChannelCode: "IN_APP_BANNER",
		Content: map[string]any{"title": "Sale"},
	}}
	sms := &stubSMS{}
	svc := application.NewIntentService(stubProfiles{}, nil, ads, sms)

	_, err := svc.IngestAndFetchAd(context.Background(), &domain.IntentProfile{
		PseudonymousID: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		IntentName:     "fashion_interest",
		Confidence:     0.9,
	}, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ads.lastSMS {
		t.Fatal("smsConsented=false must be passed through")
	}
	if sms.called {
		t.Fatal("dispatcher must not be called when sms_consented=false")
	}
}

func TestIngestAndFetchAd_SMSDispatchError(t *testing.T) {
	ads := &stubAds{ad: &application.AdSelection{
		CampaignID: "camp-sms", ChannelCode: "SMS_PLUS",
		Campaign: &campaigndomain.Campaign{ID: "camp-sms"},
	}}
	sms := &stubSMS{err: errors.New("provider down")}
	svc := application.NewIntentService(stubProfiles{}, nil, ads, sms)

	_, err := svc.IngestAndFetchAd(context.Background(), &domain.IntentProfile{
		PseudonymousID: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		IntentName:     "fashion_interest",
		Confidence:     0.9,
	}, "", true)
	if err == nil {
		t.Fatal("expected dispatch error")
	}
}

func TestIngestAndFetchAd_SMSRequiresCampaign(t *testing.T) {
	ads := &stubAds{ad: &application.AdSelection{
		CampaignID: "camp-sms", ChannelCode: "SMS_PLUS",
	}}
	sms := &stubSMS{}
	svc := application.NewIntentService(stubProfiles{}, nil, ads, sms)

	_, err := svc.IngestAndFetchAd(context.Background(), &domain.IntentProfile{
		PseudonymousID: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		IntentName:     "fashion_interest",
		Confidence:     0.9,
	}, "", true)
	if err == nil {
		t.Fatal("expected error when Campaign is nil")
	}
}
