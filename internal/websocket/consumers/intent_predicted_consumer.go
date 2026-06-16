package consumers

import (
	intentEvents "skykin-platform/internal/intents/events"
	"skykin-platform/internal/platform/messaging"
	platformWS "skykin-platform/internal/platform/websocket"
)

// IntentPredictedConsumer pushes intent prediction results to connected SDK clients.
type IntentPredictedConsumer struct {
	notifier platformWS.Notifier
}

func NewIntentPredictedConsumer(notifier platformWS.Notifier) *IntentPredictedConsumer {
	return &IntentPredictedConsumer{notifier: notifier}
}

func (c *IntentPredictedConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, intentEvents.TopicIntentPredicted, c.handle)
}

func (c *IntentPredictedConsumer) handle(e messaging.Event) {
	p, ok := e.Payload.(intentEvents.IntentPredicted)
	if !ok || p.ExternalUserID == "" || p.Intent == "" {
		return
	}
	_ = c.notifier.NotifyUser(p.ExternalUserID, map[string]any{
		"type":             "intent_predicted",
		"intent":           p.Intent,
		"confidence":       p.Confidence,
		"top_signals":      p.TopSignals,
		"reward_triggered": p.RewardTriggered,
	})
}
