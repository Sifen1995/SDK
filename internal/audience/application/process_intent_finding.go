package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	analyticsdomain "skykin-platform/internal/analytics/domain"
	"skykin-platform/internal/audience/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	outcomeMergedSegment     = "merged_segment"
	outcomeUpdatedCandidate  = "updated_candidate"
	outcomeCreatedCandidate  = "created_candidate"
	outcomeSkippedNoNewUsers = "skipped_no_new_users"
)

// FindingOutcome describes what happened when a scan finding was processed.
type FindingOutcome struct {
	Action      string
	IntentName  string
	UsersAdded  int
	SegmentID   string
	CandidateID string
}

// ProcessIntentFindingUseCase deduplicates findings against existing segments and pending candidates.
type ProcessIntentFindingUseCase struct {
	uow    domain.UnitOfWork
	logger *slog.Logger
}

func NewProcessIntentFindingUseCase(uow domain.UnitOfWork, logger *slog.Logger) *ProcessIntentFindingUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProcessIntentFindingUseCase{uow: uow, logger: logger}
}

func (uc *ProcessIntentFindingUseCase) Execute(
	ctx context.Context,
	finding analyticsdomain.IntentConsistencyFinding,
) (FindingOutcome, error) {
	users := mapFindingUsers(finding)
	now := time.Now().UTC()

	var outcome FindingOutcome
	err := uc.uow.Do(ctx, func(r domain.Repositories) error {
		seg, err := r.Segments.FindActiveByIntentSignal(ctx, finding.IntentName, now)
		switch {
		case err == nil:
			outcome, err = mergeIntoSegment(ctx, r, seg, finding.IntentName, users, uc.logger)
			return err
		case errors.Is(err, gorm.ErrRecordNotFound):
			outcome, err = upsertCandidate(ctx, r, finding, users, uc.logger)
			return err
		default:
			return err
		}
	})
	if err != nil {
		return FindingOutcome{}, err
	}
	return outcome, nil
}

func mergeIntoSegment(
	ctx context.Context,
	r domain.Repositories,
	seg *domain.AudienceSegment,
	intentName string,
	users []*domain.UserInCandidate,
	logger *slog.Logger,
) (FindingOutcome, error) {
	segUUID, err := uuid.Parse(seg.ID)
	if err != nil {
		return FindingOutcome{}, fmt.Errorf("invalid segment id: %w", err)
	}

	existing, err := r.Membership.FindPseudonymousIDsInSegment(ctx, segUUID)
	if err != nil {
		return FindingOutcome{}, err
	}
	newUsers := filterNewUsers(users, stringSet(existing))
	if len(newUsers) == 0 {
		logger.Info("no new members for existing segment", "intent", intentName, "segment_id", seg.ID)
		return FindingOutcome{Action: outcomeSkippedNoNewUsers, IntentName: intentName, SegmentID: seg.ID}, nil
	}

	if err := r.Membership.BulkInsert(ctx, segUUID, newUsers); err != nil {
		return FindingOutcome{}, err
	}
	count, err := r.Membership.CountMembers(ctx, segUUID)
	if err != nil {
		return FindingOutcome{}, err
	}
	seg.ApproximateSize = count
	if err := r.Segments.Update(ctx, seg); err != nil {
		return FindingOutcome{}, err
	}

	logger.Info("merged members into existing segment",
		"intent", intentName, "segment_id", seg.ID, "users_added", len(newUsers))
	return FindingOutcome{
		Action: outcomeMergedSegment, IntentName: intentName,
		UsersAdded: len(newUsers), SegmentID: seg.ID,
	}, nil
}

func upsertCandidate(
	ctx context.Context,
	r domain.Repositories,
	finding analyticsdomain.IntentConsistencyFinding,
	users []*domain.UserInCandidate,
	logger *slog.Logger,
) (FindingOutcome, error) {
	result, err := r.Candidates.UpsertPending(ctx, candidateFromFinding(finding), users)
	if err != nil {
		return FindingOutcome{}, err
	}
	action := outcomeUpdatedCandidate
	if result.Created {
		action = outcomeCreatedCandidate
	}
	logger.Info("pending segment candidate upserted",
		"intent", finding.IntentName, "candidate_id", result.CandidateID, "action", action)
	return FindingOutcome{
		Action: action, IntentName: finding.IntentName,
		CandidateID: result.CandidateID.String(), UsersAdded: len(users),
	}, nil
}

func mapFindingUsers(finding analyticsdomain.IntentConsistencyFinding) []*domain.UserInCandidate {
	users := make([]*domain.UserInCandidate, 0, len(finding.Users))
	for _, u := range finding.Users {
		users = append(users, &domain.UserInCandidate{
			PseudonymousID: u.PseudonymousID, Confidence: u.Confidence,
			DaysActive: u.DaysActive, LastSeenAt: u.LastSeenAt,
		})
	}
	return users
}

func candidateFromFinding(finding analyticsdomain.IntentConsistencyFinding) *domain.SegmentCandidate {
	return &domain.SegmentCandidate{
		ID: finding.FindingID, IntentName: finding.IntentName,
		UserCount: finding.UserCount, AvgConfidence: finding.AvgConfidence,
		AvgDaysActive: finding.AvgDaysActive, MinDaysActive: finding.MinDaysActive,
		LookbackDays: finding.LookbackDays, Status: domain.CandidateStatusPending,
		ScannedAt: finding.ScannedAt,
	}
}

func stringSet(ids []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

func filterNewUsers(users []*domain.UserInCandidate, existing map[string]struct{}) []*domain.UserInCandidate {
	out := make([]*domain.UserInCandidate, 0, len(users))
	for _, u := range users {
		if _, ok := existing[u.PseudonymousID]; !ok {
			out = append(out, u)
		}
	}
	return out
}
