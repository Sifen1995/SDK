package consumers

import (
	"time"

	"skykin-platform/internal/platform/messaging"
	rewardsdomain "skykin-platform/internal/rewards/domain"
)

const RewardEvaluationRequested = "rewards.evaluation.requested"

type RewardEvaluationPayload struct {
	PseudonymousID string
	IntentID       string
	IntentName     string
	Confidence     float64
	Triggered      bool
}

// RewardConsumer handles legacy impression-based reward evaluation.
type RewardConsumer struct {
	rewardRepo rewardsdomain.RewardRepository
	bus        *messaging.Bus
}

func NewRewardConsumer(rewardRepo rewardsdomain.RewardRepository, bus *messaging.Bus) *RewardConsumer {
	return &RewardConsumer{rewardRepo: rewardRepo, bus: bus}
}

func (c *RewardConsumer) Register() {
	messaging.Register(c.bus, RewardEvaluationRequested, c.handle)
}

func (c *RewardConsumer) handle(e messaging.Event) {
	p, ok := e.Payload.(RewardEvaluationPayload)
	if !ok || !p.Triggered || p.PseudonymousID == "" {
		return
	}

	rule, err := c.rewardRepo.GetRuleByIntent(e.Ctx, p.IntentName)
	if err != nil {
		return
	}

	reward := &rewardsdomain.Reward{
		PseudonymousID: p.PseudonymousID,
		IntentID:       p.IntentID,
		RuleID:         rule.ID,
		RewardType:     rule.RewardType,
		Amount:         rule.Amount,
		Currency:       rule.Currency,
		Status:         "pending",
		Message:        rule.Message,
		CreatedAt:      time.Now().UTC(),
	}

	_ = c.rewardRepo.CreateReward(e.Ctx, reward)
}
