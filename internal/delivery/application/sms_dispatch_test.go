package application

import (
	"context"
	"testing"
	"time"

	deliverydomain "skykin-platform/internal/delivery/domain"
)

type fakeSMSCampaignReader struct {
	campaign *SMSCampaign
}

func (f *fakeSMSCampaignReader) GetSMSCampaign(ctx context.Context, campaignID string) (*SMSCampaign, error) {
	_ = ctx
	_ = campaignID
	return f.campaign, nil
}

type fakeRecipientRepo struct {
	recipient *deliverydomain.DemoSMSRecipient
}

func (f *fakeRecipientRepo) FindActiveByPseudonymousID(ctx context.Context, pseudonymousID string) (*deliverydomain.DemoSMSRecipient, error) {
	_ = ctx
	_ = pseudonymousID
	return f.recipient, nil
}

type fakeAttemptRepo struct {
	attempts map[string]*deliverydomain.SMSSendAttempt
}

func (f *fakeAttemptRepo) Create(ctx context.Context, attempt *deliverydomain.SMSSendAttempt) error {
	_ = ctx
	if f.attempts == nil {
		f.attempts = map[string]*deliverydomain.SMSSendAttempt{}
	}
	attempt.ID = attempt.SendKey
	attempt.CreatedAt = time.Now().UTC()
	f.attempts[attempt.SendKey] = attempt
	return nil
}

func (f *fakeAttemptRepo) ExistsBySendKey(ctx context.Context, sendKey string) (bool, error) {
	_ = ctx
	_, ok := f.attempts[sendKey]
	return ok, nil
}

func (f *fakeAttemptRepo) FindBySendKey(ctx context.Context, sendKey string) (*deliverydomain.SMSSendAttempt, error) {
	_ = ctx
	return f.attempts[sendKey], nil
}

func (f *fakeAttemptRepo) FindByTrackingToken(ctx context.Context, trackingToken string) (*deliverydomain.SMSSendAttempt, error) {
	_ = ctx
	for _, attempt := range f.attempts {
		if attempt.TrackingToken == trackingToken {
			return attempt, nil
		}
	}
	return nil, nil
}

func (f *fakeAttemptRepo) UpdateStatus(ctx context.Context, attemptID, status, providerMessageID, errorMessage string) error {
	_ = ctx
	for _, attempt := range f.attempts {
		if attempt.ID == attemptID {
			attempt.Status = status
			attempt.ProviderMessageID = providerMessageID
			attempt.ErrorMessage = errorMessage
		}
	}
	return nil
}

func (f *fakeAttemptRepo) ListRecent(ctx context.Context, limit int) ([]deliverydomain.SMSSendAttempt, error) {
	_ = ctx
	_ = limit
	out := make([]deliverydomain.SMSSendAttempt, 0, len(f.attempts))
	for _, attempt := range f.attempts {
		out = append(out, *attempt)
	}
	return out, nil
}

type fakeDeliveryLogRepo struct {
	logs []deliverydomain.DeliveryLog
}

func (f *fakeDeliveryLogRepo) CreateBatch(ctx context.Context, logs []deliverydomain.DeliveryLog) error {
	_ = ctx
	f.logs = append(f.logs, logs...)
	return nil
}

type fakeSMSProvider struct {
	sendCount int
}

func (f *fakeSMSProvider) ProviderName() string { return "mock" }

func (f *fakeSMSProvider) Send(ctx context.Context, msg SMSMessage) (*SMSSendResult, error) {
	_ = ctx
	f.sendCount++
	return &SMSSendResult{ProviderMessageID: "provider-1"}, nil
}

func TestSMSDispatchServiceDispatchCampaignMatchCreatesAttemptAndLog(t *testing.T) {
	attempts := &fakeAttemptRepo{attempts: map[string]*deliverydomain.SMSSendAttempt{}}
	logs := &fakeDeliveryLogRepo{}
	provider := &fakeSMSProvider{}
	svc := NewSMSDispatchService(
		&fakeSMSCampaignReader{campaign: &SMSCampaign{
			ID:             "c1",
			Title:          "Deal",
			BodyText:       "Tap now",
			DestinationURL: "https://example.com",
		}},
		&fakeRecipientRepo{recipient: &deliverydomain.DemoSMSRecipient{
			UserID:    101,
			PhoneE164: "+15550000001",
			IsActive:  true,
		}},
		attempts,
		logs,
		provider,
		"http://localhost:8081",
		"test-secret",
	)

	if err := svc.DispatchCampaignMatch(context.Background(), "c1", "p1"); err != nil {
		t.Fatalf("DispatchCampaignMatch() error = %v", err)
	}
	if provider.sendCount != 1 {
		t.Fatalf("expected one provider send, got %d", provider.sendCount)
	}
	attempt, ok := attempts.attempts["c1:p1"]
	if !ok {
		t.Fatalf("expected attempt to be created")
	}
	if attempt.Status != deliverydomain.SMSSendStatusSent {
		t.Fatalf("expected status %q, got %q", deliverydomain.SMSSendStatusSent, attempt.Status)
	}
	if attempt.TrackingToken == "" {
		t.Fatalf("expected tracking token to be populated")
	}
	if len(logs.logs) != 1 || logs.logs[0].DeliveryStatus != deliverydomain.StatusDispatched {
		t.Fatalf("expected one DISPATCHED delivery log, got %+v", logs.logs)
	}
}

func TestSMSDispatchServiceDispatchCampaignMatchIsIdempotent(t *testing.T) {
	attempts := &fakeAttemptRepo{attempts: map[string]*deliverydomain.SMSSendAttempt{}}
	provider := &fakeSMSProvider{}
	svc := NewSMSDispatchService(
		&fakeSMSCampaignReader{campaign: &SMSCampaign{
			ID:             "c1",
			BodyText:       "Tap now",
			DestinationURL: "https://example.com",
		}},
		&fakeRecipientRepo{recipient: &deliverydomain.DemoSMSRecipient{
			UserID:    101,
			PhoneE164: "+15550000001",
			IsActive:  true,
		}},
		attempts,
		&fakeDeliveryLogRepo{},
		provider,
		"http://localhost:8081",
		"test-secret",
	)

	if err := svc.DispatchCampaignMatch(context.Background(), "c1", "p1"); err != nil {
		t.Fatalf("first DispatchCampaignMatch() error = %v", err)
	}
	if err := svc.DispatchCampaignMatch(context.Background(), "c1", "p1"); err != nil {
		t.Fatalf("second DispatchCampaignMatch() error = %v", err)
	}
	if provider.sendCount != 1 {
		t.Fatalf("expected one provider send after duplicate match, got %d", provider.sendCount)
	}
}
