package events

// Internal application event topic names (not SDK event types).
const (
	TopicEventReceived             = "events.event_received"
	TopicEventStored               = "events.event_stored"
	TopicIntentEvaluationRequested = "events.intent_evaluation_requested"
)

// EventReceived is published after SDK events are accepted into the platform.
type EventReceived struct {
	EventID        string
	PseudonymousID string
	ApplicationID  string
	EventType      string
	Domain         string
	SessionID      string
}

// EventStored is published after events are persisted.
type EventStored struct {
	EventID        string
	PseudonymousID string
	ApplicationID  string
	EventType      string
	Domain         string
	SessionID      string
}

// IntentEvaluationRequested signals downstream intent workflows.
type IntentEvaluationRequested struct {
	PseudonymousID string
	ApplicationID  string
	SessionID      string
	EventCount     int
}
