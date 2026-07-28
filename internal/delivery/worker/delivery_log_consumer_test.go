package worker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	deliverydomain "skykin-platform/internal/delivery/domain"
	platformredis "skykin-platform/internal/platform/redis"
)

func TestMapStreamToDeliveryLog_ValidInstallToken(t *testing.T) {
	t.Setenv("CLICK_TOKEN_SECRET", "super-secret")

	campaignID := "550e8400-e29b-41d4-a716-446655440000"
	token := buildTestInstallToken(t, "super-secret", campaignID)

	msg := platformredis.StreamMessage{
		Values: map[string]string{
			"campaign_id":   campaignID,
			"event_type":    "install",
			"install_token": token,
		},
	}

	got, err := mapStreamToDeliveryLog(msg)
	if err != nil {
		t.Fatalf("mapStreamToDeliveryLog returned error: %v", err)
	}
	if got == nil {
		t.Fatal("mapStreamToDeliveryLog returned nil log")
	}
	if got.DeliveryStatus != deliverydomain.StatusConverted {
		t.Fatalf("expected %q delivery status, got %q", deliverydomain.StatusConverted, got.DeliveryStatus)
	}
}

func TestMapStreamToDeliveryLog_InvalidInstallToken(t *testing.T) {
	t.Setenv("CLICK_TOKEN_SECRET", "super-secret")

	campaignID := "550e8400-e29b-41d4-a716-446655440000"
	msg := platformredis.StreamMessage{
		Values: map[string]string{
			"campaign_id":   campaignID,
			"event_type":    "install",
			"install_token": "bad-token",
		},
	}

	if _, err := mapStreamToDeliveryLog(msg); err == nil {
		t.Fatal("expected invalid install token error")
	}
}

func buildTestInstallToken(t *testing.T, secret, campaignID string) string {
	t.Helper()
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(campaignID))
	return hex.EncodeToString(h.Sum(nil))
}
