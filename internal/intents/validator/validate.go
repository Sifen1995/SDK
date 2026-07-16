package validator

import (
	"fmt"
	"strings"

	"skykin-platform/internal/intents/domain"
)

// ValidateIntentProfile checks the SDK intent profile before persistence.
func ValidateIntentProfile(profile *domain.IntentProfile) error {
	if profile == nil {
		return fmt.Errorf("intent profile is required")
	}
	if strings.TrimSpace(profile.PseudonymousID) == "" {
		return fmt.Errorf("pseudonymous_id is required")
	}
	if strings.TrimSpace(profile.IntentName) == "" {
		return fmt.Errorf("intent_name is required")
	}
	if profile.Confidence < 0.0 || profile.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}
	return nil
}
