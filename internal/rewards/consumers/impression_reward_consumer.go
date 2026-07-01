package consumers

import (
	"time"

	rewardsdomain "skykin-platform/internal/rewards/domain"
	usersdomain "skykin-platform/internal/users/domain"
	"skykin-platform/internal/platform/messaging"
	wsConsumers "skykin-platform/internal/websocket/consumers"
)

const RewardEvaluationRequested = "rewards.evaluation.requested"

type RewardEvaluationPayload struct {
	ExternalUserID string
	IntentID       string
	IntentName     string
	Confidence     float64
	Triggered      bool
}

type RewardConsumer struct {
	rewardRepo rewardsdomain.RewardRepository
	userRepo   usersdomain.UserRepository
	bus        *messaging.Bus
}

func NewRewardConsumer(rewardRepo rewardsdomain.RewardRepository, userRepo usersdomain.UserRepository, bus *messaging.Bus) *RewardConsumer {
	return &RewardConsumer{rewardRepo: rewardRepo, userRepo: userRepo, bus: bus}
}

func (c *RewardConsumer) Register() {
	messaging.Register(c.bus, RewardEvaluationRequested, c.handle)
}

func (c *RewardConsumer) handle(e messaging.Event) {
	p, ok := e.Payload.(RewardEvaluationPayload)
	if !ok || !p.Triggered {
		return
	}

	rule, err := c.rewardRepo.GetRuleByIntent(e.Ctx, p.IntentName)
	if err != nil {
		return
	}

	user, err := c.userRepo.FindOrCreate(e.Ctx, p.ExternalUserID)
	if err != nil {
		return
	}

	reward := &rewardsdomain.Reward{
		UserID:     user.ID,
		IntentID:   p.IntentID,
		RuleID:     rule.ID,
		RewardType: rule.RewardType,
		Amount:     rule.Amount,
		Currency:   rule.Currency,
		Status:     "pending",
		Message:    rule.Message,
		CreatedAt:  time.Now().UTC(),
	}

	if err := c.rewardRepo.CreateReward(e.Ctx, reward); err != nil {
		return
	}

	c.bus.Publish(messaging.Event{
		Name: wsConsumers.RewardCreatedEvent,
		Ctx:  e.Ctx,
		Payload: wsConsumers.RewardCreatedPayload{
			ExternalUserID: p.ExternalUserID,
			IntentName:     p.IntentName,
			Confidence:     p.Confidence,
			Reward:         reward,
		},
	})
}
