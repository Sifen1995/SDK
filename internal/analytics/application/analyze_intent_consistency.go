package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"skykin-platform/internal/analytics/domain"

	"github.com/google/uuid"
)

// IntentConsistencyReader loads users with sustained intent signals.
type IntentConsistencyReader interface {
	FindConsistentUsers(ctx context.Context, intentName string, minConf float64, lookbackDays, minDays, maxAgeDays int) ([]*domain.ConsistentUser, error)
}

type AnalyzeIntentConsistencyUseCase struct {
	intentRepo IntentConsistencyReader
	config     domain.ClassificationConfig
	processor  IntentFindingProcessor
	logger     *slog.Logger
}

func NewAnalyzeIntentConsistencyUseCase(
	intentRepo IntentConsistencyReader,
	config domain.ClassificationConfig,
	processor IntentFindingProcessor,
	logger *slog.Logger,
) *AnalyzeIntentConsistencyUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &AnalyzeIntentConsistencyUseCase{
		intentRepo: intentRepo, config: config, processor: processor, logger: logger,
	}
}

func (uc *AnalyzeIntentConsistencyUseCase) Run(ctx context.Context) (*RunReport, error) {
	report := &RunReport{}
	for _, intentClass := range uc.config.IntentClasses {
		outcome, err := uc.scanIntent(ctx, intentClass)
		if err != nil {
			uc.logger.Error("intent scan failed", "intent", intentClass, "error", err)
			report.IntentsFailed = append(report.IntentsFailed, IntentScanFailure{
				IntentName: intentClass,
				Error:      err.Error(),
			})
			continue
		}
		if outcome == nil {
			continue
		}
		aggregateReport(report, *outcome)
	}
	// Every intent class failed: the scan produced no usable result at all.
	if len(report.IntentsFailed) == len(uc.config.IntentClasses) && len(uc.config.IntentClasses) > 0 {
		return nil, fmt.Errorf("intent consistency scan failed for all %d intent classes: %s",
			len(report.IntentsFailed), report.IntentsFailed[0].Error)
	}
	report.Partial = len(report.IntentsFailed) > 0
	report.Message = buildRunMessage(report)
	return report, nil
}

func (uc *AnalyzeIntentConsistencyUseCase) scanIntent(
	ctx context.Context,
	intentClass string,
) (*FindingProcessResult, error) {
	users, err := uc.intentRepo.FindConsistentUsers(ctx, intentClass,
		uc.config.MinConfidence, uc.config.LookbackDays,
		uc.config.MinDaysActive, uc.config.MaxAgeDays)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		uc.logger.Info("no consistent users", "intent", intentClass)
		return nil, nil
	}

	finding := buildFinding(intentClass, users, uc.config)
	if uc.processor == nil {
		uc.logger.Warn("no finding processor wired", "intent", intentClass)
		return nil, nil
	}
	result, err := uc.processor.Process(ctx, finding)
	if err != nil {
		return nil, err
	}
	uc.logger.Info("intent finding processed",
		"intent", intentClass, "action", result.Action, "users_added", result.UsersAdded)
	return &result, nil
}

func buildFinding(intentClass string, users []*domain.ConsistentUser, cfg domain.ClassificationConfig) domain.IntentConsistencyFinding {
	var sumConf, sumDays float64
	for _, u := range users {
		sumConf += u.Confidence
		sumDays += float64(u.DaysActive)
	}
	n := float64(len(users))
	return domain.IntentConsistencyFinding{
		FindingID: uuid.New(), IntentName: intentClass, Users: users,
		UserCount: len(users), AvgConfidence: sumConf / n, AvgDaysActive: sumDays / n,
		MinDaysActive: cfg.MinDaysActive, LookbackDays: cfg.LookbackDays,
		ScannedAt: time.Now().UTC(),
	}
}

func aggregateReport(report *RunReport, outcome FindingProcessResult) {
	switch outcome.Action {
	case "created_candidate":
		report.CandidatesCreated++
	case "updated_candidate":
		report.CandidatesUpdated++
	case "merged_segment":
		report.SegmentsEnriched++
		report.UsersAdded += outcome.UsersAdded
	case "skipped_no_new_users":
		report.IntentsSkipped = append(report.IntentsSkipped, outcome.IntentName)
	}
}

func buildRunMessage(report *RunReport) string {
	prefix := "Scan complete."
	if report.Partial {
		prefix = fmt.Sprintf("Scan incomplete: %d intent class(es) failed.", len(report.IntentsFailed))
	}
	if report.CandidatesCreated == 0 && report.CandidatesUpdated == 0 {
		return prefix + " No new segment candidates."
	}
	return fmt.Sprintf("%s %d new candidate(s), %d updated.",
		prefix, report.CandidatesCreated, report.CandidatesUpdated)
}
