package consumers

import (
	"skykin-platform/internal/platform/messaging"
	platformWS "skykin-platform/internal/platform/websocket"
	rewardsdomain "skykin-platform/internal/rewards/domain"
)

const RewardCreatedEvent = "rewards.created"

type RewardCreatedPayload struct {
	ExternalUserID string
	IntentName     string
	Confidence     float64
	Reward         *rewardsdomain.Reward
}

type RewardCreatedConsumer struct {
	bus      *messaging.Bus
	notifier platformWS.Notifier
}

func NewRewardCreatedConsumer(bus *messaging.Bus, notifier platformWS.Notifier) *RewardCreatedConsumer {
	return &RewardCreatedConsumer{bus: bus, notifier: notifier}
}

func (c *RewardCreatedConsumer) Register() {
	messaging.Register(c.bus, RewardCreatedEvent, c.handle)
}

func (c *RewardCreatedConsumer) handle(e messaging.Event) {
	p, ok := e.Payload.(RewardCreatedPayload)
	if !ok || p.Reward == nil {
		return
	}
	_ = c.notifier.NotifyUser(p.ExternalUserID, map[string]any{
		"type":        "reward_earned",
		"reward_id":   p.Reward.ID,
		"reward_type": p.Reward.RewardType,
		"amount":      p.Reward.Amount,
		"currency":    p.Reward.Currency,
		"message":     p.Reward.Message,
		"intent":      p.IntentName,
		"confidence":  p.Confidence,
		"created_at":  p.Reward.CreatedAt,
	})
}
