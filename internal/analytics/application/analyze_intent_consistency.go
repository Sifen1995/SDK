package application

import (
	"context"
	"log/slog"
	"time"

	"skykin-platform/internal/analytics/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
)

// IntentConsistencyReader loads users with sustained intent signals.
type IntentConsistencyReader interface {
	FindConsistentUsers(ctx context.Context, intentName string, minConf float64, lookbackDays, minDays, maxAgeDays int) ([]*domain.ConsistentUser, error)
}

type AnalyzeIntentConsistencyUseCase struct {
	intentRepo IntentConsistencyReader
	config     domain.ClassificationConfig
	bus        *messaging.Bus
	logger     *slog.Logger
}

func NewAnalyzeIntentConsistencyUseCase(
	intentRepo IntentConsistencyReader,
	config domain.ClassificationConfig,
	bus *messaging.Bus,
	logger *slog.Logger,
) *AnalyzeIntentConsistencyUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &AnalyzeIntentConsistencyUseCase{
		intentRepo: intentRepo, config: config, bus: bus, logger: logger,
	}
}

func (uc *AnalyzeIntentConsistencyUseCase) Run(ctx context.Context) error {
	for _, intentClass := range uc.config.IntentClasses {
		users, err := uc.intentRepo.FindConsistentUsers(ctx, intentClass,
			uc.config.MinConfidence, uc.config.LookbackDays,
			uc.config.MinDaysActive, uc.config.MaxAgeDays)
		if err != nil {
			uc.logger.Error("consistent users query failed", "intent", intentClass, "error", err)
			continue
		}
		if len(users) == 0 {
			uc.logger.Info("no consistent users", "intent", intentClass)
			continue
		}
		var sumConf, sumDays float64
		for _, u := range users {
			sumConf += u.Confidence
			sumDays += float64(u.DaysActive)
		}
		n := float64(len(users))
		finding := domain.IntentConsistencyFinding{
			FindingID: uuid.New(), IntentName: intentClass, Users: users,
			UserCount: len(users), AvgConfidence: sumConf / n, AvgDaysActive: sumDays / n,
			MinDaysActive: uc.config.MinDaysActive, LookbackDays: uc.config.LookbackDays,
			ScannedAt: time.Now().UTC(),
		}
		uc.bus.Publish(messaging.Event{
			Name: domain.TopicIntentConsistencyFinding, Payload: finding, Ctx: ctx,
		})
		uc.logger.Info("intent consistency finding published", "intent", intentClass,
			"user_count", finding.UserCount, "avg_confidence", finding.AvgConfidence,
			"avg_days_active", finding.AvgDaysActive)
	}
	return nil
}
