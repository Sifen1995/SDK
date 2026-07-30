package infrastructure

import (
	"context"
	"testing"

	deliveryapp "skykin-platform/internal/delivery/application"
)

func TestMockSMSProviderSendIsDeterministic(t *testing.T) {
	provider := NewMockSMSProvider("")
	msg := deliveryapp.SMSMessage{To: "+15550000001", Body: "hello"}

	first, err := provider.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("first Send() error = %v", err)
	}
	second, err := provider.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("second Send() error = %v", err)
	}
	if first.ProviderMessageID == "" || first.ProviderMessageID != second.ProviderMessageID {
		t.Fatalf("expected deterministic provider message id, got %q and %q", first.ProviderMessageID, second.ProviderMessageID)
	}
}
