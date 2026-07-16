package bootstrap

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"skykin-platform/configs"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	analyticsInfra "skykin-platform/internal/analytics/infrastructure"
	audienceApp "skykin-platform/internal/audience/application"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	intentsInfra "skykin-platform/internal/intents/infrastructure"
	platformredis "skykin-platform/internal/platform/redis"

	"gorm.io/gorm"
)

// IntentConsistencyJobs holds the analysis use case and shared audience candidate repo.
type IntentConsistencyJobs struct {
	AnalyzeUC     *analyticsApp.AnalyzeIntentConsistencyUseCase
	CandidateRepo *audienceInfra.CandidateRepository
}

// SetupIntentConsistency wires analysis and synchronous finding processing in audience.
func SetupIntentConsistency(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) *IntentConsistencyJobs {
	if logger == nil {
		logger = slog.Default()
	}
	candidateRepo := audienceInfra.NewCandidateRepository(db)
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	membershipRepo := audienceInfra.NewMembershipRepository(db)
	intentRepo := intentsInfra.NewConsistencyReader(db, cfg)

	processUC := audienceApp.NewProcessIntentFindingUseCase(
		segmentRepo, membershipRepo, candidateRepo, logger,
	)
	processor := audienceApp.NewFindingProcessorAdapter(processUC)

	analyzeUC := analyticsApp.NewAnalyzeIntentConsistencyUseCase(
		intentRepo, analyticsdomain.DefaultConfig(), processor, logger,
	)

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
		if _, err := jobs.AnalyzeUC.Run(context.Background()); err != nil {
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

// StartAnalyticsAggregateWorker drains queue:analytics_aggregate into Postgres upserts.
func StartAnalyticsAggregateWorker(db *gorm.DB, cfg *configs.Config, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	addr := strings.TrimSpace(cfg.RedisAddr)
	if addr == "" {
		return
	}
	rdb, err := platformredis.NewRedisClient(addr)
	if err != nil {
		logger.Warn("analytics aggregate worker: redis unavailable", "error", err)
		return
	}
	analyticsInfra.StartAnalyticsAggregateWorker(db, rdb, logger)
}
