package application

import (
	"context"
	"log/slog"

	analyticsdomain "skykin-platform/internal/analytics/domain"
	"skykin-platform/internal/audience/domain"
)

type SaveCandidateFromFindingUseCase struct {
	repo   domain.CandidateRepository
	logger *slog.Logger
}

func NewSaveCandidateFromFindingUseCase(repo domain.CandidateRepository, logger *slog.Logger) *SaveCandidateFromFindingUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &SaveCandidateFromFindingUseCase{repo: repo, logger: logger}
}

func (uc *SaveCandidateFromFindingUseCase) Execute(ctx context.Context, finding analyticsdomain.IntentConsistencyFinding) error {
	users := make([]*domain.UserInCandidate, 0, len(finding.Users))
	for _, u := range finding.Users {
		users = append(users, &domain.UserInCandidate{
			UserID: u.UserID, Confidence: u.Confidence,
			DaysActive: u.DaysActive, LastSeenAt: u.LastSeenAt,
		})
	}
	candidate := &domain.SegmentCandidate{
		ID: finding.FindingID, IntentName: finding.IntentName,
		UserCount: finding.UserCount, AvgConfidence: finding.AvgConfidence,
		AvgDaysActive: finding.AvgDaysActive, MinDaysActive: finding.MinDaysActive,
		LookbackDays: finding.LookbackDays, Status: domain.CandidateStatusPending,
		ScannedAt: finding.ScannedAt,
	}
	if err := uc.repo.Save(ctx, candidate, users); err != nil {
		return err
	}
	uc.logger.Info("segment candidate saved from finding", "intent", finding.IntentName, "candidate_id", finding.FindingID)
	return nil
}

type ListSegmentCandidatesUseCase struct {
	repo domain.CandidateRepository
}

func NewListSegmentCandidatesUseCase(repo domain.CandidateRepository) *ListSegmentCandidatesUseCase {
	return &ListSegmentCandidatesUseCase{repo: repo}
}

func (uc *ListSegmentCandidatesUseCase) ListByStatus(ctx context.Context, status domain.CandidateStatus) ([]*domain.SegmentCandidate, error) {
	return uc.repo.FindByStatus(ctx, status)
}
