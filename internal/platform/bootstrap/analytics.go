package bootstrap

import (
	"context"
	"log/slog"
	"time"

	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	audienceApp "skykin-platform/internal/audience/application"
	audienceConsumers "skykin-platform/internal/audience/consumers"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	"skykin-platform/configs"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	"skykin-platform/internal/platform/messaging"

	"gorm.io/gorm"
)

// IntentConsistencyJobs holds the analysis use case and shared audience candidate repo.
type IntentConsistencyJobs struct {
	AnalyzeUC     *analyticsApp.AnalyzeIntentConsistencyUseCase
	CandidateRepo *audienceInfra.CandidateRepository
}

// SetupIntentConsistency wires analysis, event publishing, and audience consumer.
func SetupIntentConsistency(db *gorm.DB, cfg *configs.Config, bus *messaging.Bus, logger *slog.Logger) *IntentConsistencyJobs {
	if logger == nil {
		logger = slog.Default()
	}
	candidateRepo := audienceInfra.NewCandidateRepository(db)
	intentRepo := intentsInfra.NewConsistencyReader(db, cfg)

	analyzeUC := analyticsApp.NewAnalyzeIntentConsistencyUseCase(
		intentRepo, analyticsdomain.DefaultConfig(), bus, logger,
	)
	saveUC := audienceApp.NewSaveCandidateFromFindingUseCase(candidateRepo, logger)
	audienceConsumers.NewFindingConsumer(saveUC, logger).Register(bus)

	return &IntentConsistencyJobs{AnalyzeUC: analyzeUC, CandidateRepo: candidateRepo}
}

// StartIntentConsistencyJobs runs analysis on a 24h ticker (plus once after startup delay).
func StartIntentConsistencyJobs(jobs *IntentConsistencyJobs, logger *slog.Logger) {
	if jobs == nil || jobs.AnalyzeUC == nil {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	run := func() {
		if err := jobs.AnalyzeUC.Run(context.Background()); err != nil {
			logger.Error("intent consistency analysis failed", "error", err)
		}
	}
	time.AfterFunc(5*time.Second, run)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			run()
		}
	}()
	logger.Info("intent consistency analysis scheduled, interval: 24h")
}
