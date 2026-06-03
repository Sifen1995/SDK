package events

// Internal application event topic names (not SDK event types).
const (
	TopicEventReceived             = "events.event_received"
	TopicEventStored               = "events.event_stored"
	TopicIntentEvaluationRequested = "events.intent_evaluation_requested"
)

// EventReceived is published after SDK events are accepted into the platform.
type EventReceived struct {
	EventID       string
	UserID        string
	ApplicationID string
	EventType     string
	Domain        string
	SessionID     string
}

// EventStored is published after events are persisted.
type EventStored struct {
	EventID       string
	UserID        string
	ApplicationID string
	EventType     string
	Domain        string
	SessionID     string
}

// IntentEvaluationRequested signals downstream intent workflows.
type IntentEvaluationRequested struct {
	UserID         string // SDK external user id
	InternalUserID string // users.id UUID for ML feature queries
	ApplicationID  string
	SessionID      string
	EventCount     int
}
