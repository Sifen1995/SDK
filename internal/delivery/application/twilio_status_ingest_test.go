package application

import (
	"context"
	"net/url"
	"testing"

	deliverydomain "skykin-platform/internal/delivery/domain"
)

type fakeStatusAttemptRepo struct {
	attempt *deliverydomain.SMSSendAttempt
}

func (f *fakeStatusAttemptRepo) Create(ctx context.Context, attempt *deliverydomain.SMSSendAttempt) error {
	_ = ctx
	f.attempt = attempt
	return nil
}

func (f *fakeStatusAttemptRepo) ExistsBySendKey(ctx context.Context, sendKey string) (bool, error) {
	_ = ctx
	_ = sendKey
	return false, nil
}

func (f *fakeStatusAttemptRepo) FindBySendKey(ctx context.Context, sendKey string) (*deliverydomain.SMSSendAttempt, error) {
	_ = ctx
	_ = sendKey
	return f.attempt, nil
}

func (f *fakeStatusAttemptRepo) FindByTrackingToken(ctx context.Context, trackingToken string) (*deliverydomain.SMSSendAttempt, error) {
	_ = ctx
	_ = trackingToken
	return f.attempt, nil
}

func (f *fakeStatusAttemptRepo) UpdateStatus(ctx context.Context, attemptID, status, providerMessageID, errorMessage string) error {
	_ = ctx
	_ = attemptID
	f.attempt.Status = status
	f.attempt.ProviderMessageID = providerMessageID
	f.attempt.ErrorMessage = errorMessage
	return nil
}

func (f *fakeStatusAttemptRepo) ListRecent(ctx context.Context, limit int) ([]deliverydomain.SMSSendAttempt, error) {
	_ = ctx
	_ = limit
	return nil, nil
}

func TestTwilioStatusIngestServiceProcessStatusUpdatesAttempt(t *testing.T) {
	repo := &fakeStatusAttemptRepo{attempt: &deliverydomain.SMSSendAttempt{ID: "a1", Status: deliverydomain.SMSSendStatusSent}}
	svc := NewTwilioStatusIngestService(repo, "secret")

	if err := svc.ProcessStatus(context.Background(), "send-key", "SM123", "delivered"); err != nil {
		t.Fatalf("ProcessStatus() error = %v", err)
	}
	if repo.attempt.Status != deliverydomain.SMSSendStatusDelivered {
		t.Fatalf("expected delivered status, got %q", repo.attempt.Status)
	}
	if repo.attempt.ProviderMessageID != "SM123" {
		t.Fatalf("expected provider message id to be updated")
	}
}

func TestTwilioStatusIngestServiceVerifySignatureRejectsMissingSignature(t *testing.T) {
	svc := NewTwilioStatusIngestService(&fakeStatusAttemptRepo{}, "secret")
	ok := svc.VerifySignature("https://example.com/status", url.Values{"MessageStatus": {"sent"}}, "")
	if ok {
		t.Fatalf("expected VerifySignature() to reject missing signature")
	}
}
