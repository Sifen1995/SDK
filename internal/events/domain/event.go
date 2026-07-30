package domain

import "time"

// EventType represents a generic behavioral signal. Meaning is derived from
// domain + metadata, not from product-specific columns.
type EventType string

const (
	EventTypeSessionStarted       EventType = "session_started"
	EventTypeScreenViewed         EventType = "screen_viewed"
	EventTypeContentViewed        EventType = "content_viewed"
	EventTypeSearchPerformed      EventType = "search_performed"
	EventTypeInteractionReceived  EventType = "interaction_received"
	EventTypeScrollActivity       EventType = "scroll_activity"
	EventTypeNotificationOpened   EventType = "notification_opened"
	EventTypeCampaignImpression   EventType = "campaign_impression"
	EventTypeCampaignClicked      EventType = "campaign_clicked"
	EventTypeConversionCompleted  EventType = "conversion_completed"
	EventTypeTransactionCompleted EventType = "transaction_completed"
	EventTypeRewardClaimed        EventType = "reward_claimed"
)

// ValidEventTypes lists allowed SDK event types for validation.
var ValidEventTypes = []EventType{
	EventTypeSessionStarted,
	EventTypeScreenViewed,
	EventTypeContentViewed,
	EventTypeSearchPerformed,
	EventTypeInteractionReceived,
	EventTypeScrollActivity,
	EventTypeNotificationOpened,
	EventTypeCampaignImpression,
	EventTypeCampaignClicked,
	EventTypeConversionCompleted,
	EventTypeTransactionCompleted,
	EventTypeRewardClaimed,
}

func (t EventType) IsValid() bool {
	for _, allowed := range ValidEventTypes {
		if t == allowed {
			return true
		}
	}
	return false
}

// Event is the core domain entity for behavioral ingestion.
type Event struct {
	ID             string
	EventID        string
	PseudonymousID string
	ApplicationID  string
	SessionID     string
	EventType     EventType
	Domain        string
	ScreenName    string
	Metadata      map[string]any
	DeviceType    string
	Platform      string
	AppVersion    string
	CreatedAt     time.Time
}
