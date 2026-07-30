package application_test

import (
	"context"
	"errors"
	"testing"

	"skykin-platform/internal/consent/application"
	"skykin-platform/internal/consent/domain"
	"skykin-platform/internal/platform/messaging"

	"gorm.io/gorm"
)

type syncBus struct {
	events []messaging.Event
}

func (b *syncBus) PublishSync(event messaging.Event) {
	b.events = append(b.events, event)
}

type stubDemoLookup struct {
	id  string
	err error
}

func (s stubDemoLookup) FindOneDemoPseudonymousID(ctx context.Context) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.id, nil
}

func TestCreateConsent_SMSConsentedFalse_StartsSaga(t *testing.T) {
	bus := &syncBus{}
	uc := application.NewCreateConsentUseCase(bus, stubDemoLookup{id: "demo-id"}, nil)

	result, err := uc.Execute(context.Background(), application.CreateConsentCommand{
		ConsentLevel: "individual",
		SMSConsented: false,
		SDKVersion:   "1.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SMSConsented {
		t.Fatal("expected SMSConsented false")
	}
	if len(bus.events) != 1 {
		t.Fatalf("expected saga event, got %d", len(bus.events))
	}
	payload, ok := bus.events[0].Payload.(domain.ConsentRegistrationRequestedPayload)
	if !ok || payload.SMSConsented {
		t.Fatal("expected saga payload with sms_consented=false")
	}
	if result.PseudonymousID == "demo-id" {
		t.Fatal("non-SMS path must not reuse demo id")
	}
}

func TestCreateConsent_SMSConsentedTrue_ReturnsDemoMapping(t *testing.T) {
	bus := &syncBus{}
	uc := application.NewCreateConsentUseCase(bus, stubDemoLookup{
		id: "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
	}, nil)

	result, err := uc.Execute(context.Background(), application.CreateConsentCommand{
		ConsentLevel: "individual",
		SMSConsented: true,
		SDKVersion:   "1.0.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.SMSConsented {
		t.Fatal("expected SMSConsented true")
	}
	if result.PseudonymousID != "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d" {
		t.Fatalf("expected demo id, got %s", result.PseudonymousID)
	}
	if len(bus.events) != 0 {
		t.Fatal("sms_consented=true must not start provisioning saga")
	}
}

func TestCreateConsent_SMSConsentedTrue_NoDemoUsers(t *testing.T) {
	bus := &syncBus{}
	uc := application.NewCreateConsentUseCase(bus, stubDemoLookup{err: gorm.ErrRecordNotFound}, nil)

	_, err := uc.Execute(context.Background(), application.CreateConsentCommand{
		ConsentLevel: "individual",
		SMSConsented: true,
		SDKVersion:   "1.0.0",
	})
	if err == nil {
		t.Fatal("expected error when no demo users")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) && err.Error() != "no demo users available for sms_consented consent" {
		t.Fatalf("unexpected error: %v", err)
	}
}
