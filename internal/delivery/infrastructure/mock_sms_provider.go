package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	deliveryapp "skykin-platform/internal/delivery/application"
)

type MockSMSProvider struct {
	failFor string
}

func NewMockSMSProvider(failFor string) *MockSMSProvider {
	return &MockSMSProvider{failFor: strings.TrimSpace(failFor)}
}

func (p *MockSMSProvider) ProviderName() string { return "mock" }

func (p *MockSMSProvider) Send(ctx context.Context, msg deliveryapp.SMSMessage) (*deliveryapp.SMSSendResult, error) {
	_ = ctx
	if p.failFor != "" && strings.Contains(msg.To, p.failFor) {
		return nil, fmt.Errorf("mock sms send failure for %s", msg.To)
	}
	sum := sha256.Sum256([]byte(msg.To + "|" + msg.Body))
	return &deliveryapp.SMSSendResult{
		ProviderMessageID: "mock-" + hex.EncodeToString(sum[:8]),
	}, nil
}
