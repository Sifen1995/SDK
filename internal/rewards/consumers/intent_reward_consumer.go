package consumers

import (
	"context"
	"log/slog"
	"time"

	intentEvents "skykin-platform/internal/intents/events"
	"skykin-platform/internal/platform/messaging"
	wsConsumers "skykin-platform/internal/websocket/consumers"
	"skykin-platform/internal/rewards/infrastructure"
	"skykin-platform/internal/rewards/model"
)

// IntentRewardConsumer creates reward rows when intent prediction qualifies.
type IntentRewardConsumer struct {
	rewards infrastructure.RewardRepository
	bus     *messaging.Bus
	log     *slog.Logger
}

func NewIntentRewardConsumer(rewards infrastructure.RewardRepository, bus *messaging.Bus, log *slog.Logger) *IntentRewardConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &IntentRewardConsumer{rewards: rewards, bus: bus, log: log}
}

func (c *IntentRewardConsumer) Register(bus *messaging.Bus) {
	messaging.Register(bus, intentEvents.TopicIntentRewardEligible, c.handle)
}

func (c *IntentRewardConsumer) handle(e messaging.Event) {
	evt, ok := e.Payload.(intentEvents.IntentRewardEligible)
	if !ok {
		c.log.Error("invalid intent reward eligible payload")
		return
	}
	ctx := e.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	rule, err := c.rewards.GetRuleByIntent(ctx, evt.IntentName)
	if err != nil {
		c.log.Warn("no reward rule for intent", "intent", evt.IntentName, "error", err)
		return
	}

	reward := &model.Reward{
		UserID:     evt.InternalUserID,
		IntentID:   evt.IntentID,
		RuleID:     rule.ID,
		RewardType: rule.RewardType,
		Amount:     rule.Amount,
		Currency:   rule.Currency,
		Status:     "pending",
		Message:    rule.Message,
		CreatedAt:  time.Now().UTC(),
	}
	if err := c.rewards.CreateReward(ctx, reward); err != nil {
		c.log.Warn("create reward failed", "user_id", evt.ExternalUserID, "error", err)
		return
	}

	if c.bus != nil {
		c.bus.Publish(messaging.Event{
			Name: wsConsumers.RewardCreatedEvent,
			Ctx:  ctx,
			Payload: wsConsumers.RewardCreatedPayload{
				ExternalUserID: evt.ExternalUserID,
				IntentName:     evt.IntentName,
				Reward:         reward,
			},
		})
	}
}
