package validation

import (
	"fmt"
	"strings"

	"skykin-platform/internal/events/domain"

	"github.com/google/uuid"
)

// EventInput validates a single SDK event payload.
func EventInput(
	eventID, eventType, domainName, sessionID, screenName string,
	metadata map[string]any,
) error {
	if strings.TrimSpace(eventID) == "" {
		return fmt.Errorf("event_id is required")
	}
	if _, err := uuid.Parse(eventID); err != nil {
		return fmt.Errorf("event_id must be a valid UUID")
	}

	et := domain.EventType(strings.TrimSpace(eventType))
	if !et.IsValid() {
		return fmt.Errorf("unsupported event_type: %s", eventType)
	}

	if strings.TrimSpace(domainName) == "" {
		return fmt.Errorf("domain is required")
	}

	if metadata == nil {
		return fmt.Errorf("metadata is required")
	}

	if sessionID != "" {
		if _, err := uuid.Parse(sessionID); err != nil {
			return fmt.Errorf("session_id must be a valid UUID when provided")
		}
	}

	_ = screenName
	return nil
}
