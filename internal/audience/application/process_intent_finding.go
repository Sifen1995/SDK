package application

import (
	"context"
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
	segments   domain.SegmentRepository
	membership domain.MembershipRepository
	candidates domain.CandidateRepository
	logger     *slog.Logger
}

func NewProcessIntentFindingUseCase(
	segments domain.SegmentRepository,
	membership domain.MembershipRepository,
	candidates domain.CandidateRepository,
	logger *slog.Logger,
) *ProcessIntentFindingUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProcessIntentFindingUseCase{
		segments: segments, membership: membership, candidates: candidates, logger: logger,
	}
}

func (uc *ProcessIntentFindingUseCase) Execute(
	ctx context.Context,
	finding analyticsdomain.IntentConsistencyFinding,
) (FindingOutcome, error) {
	users := mapFindingUsers(finding)
	now := time.Now().UTC()

	if seg, err := uc.segments.FindActiveByIntentSignal(ctx, finding.IntentName, now); err == nil {
		return uc.mergeIntoSegment(ctx, seg, finding.IntentName, users)
	} else if err != gorm.ErrRecordNotFound {
		return FindingOutcome{}, err
	}

	if pending, err := uc.candidates.FindPendingByIntentName(ctx, finding.IntentName); err == nil {
		return uc.refreshPendingCandidate(ctx, pending.ID, finding, users)
	} else if err != gorm.ErrRecordNotFound {
		return FindingOutcome{}, err
	}

	return uc.createCandidate(ctx, finding, users)
}

func (uc *ProcessIntentFindingUseCase) mergeIntoSegment(
	ctx context.Context,
	seg *domain.AudienceSegment,
	intentName string,
	users []*domain.UserInCandidate,
) (FindingOutcome, error) {
	segUUID, err := uuid.Parse(seg.ID)
	if err != nil {
		return FindingOutcome{}, fmt.Errorf("invalid segment id: %w", err)
	}

	existing, err := uc.membership.FindUsersInSegment(ctx, segUUID)
	if err != nil {
		return FindingOutcome{}, err
	}
	memberSet := uuidSet(existing)
	newUsers := filterNewUsers(users, memberSet)
	if len(newUsers) == 0 {
		uc.logger.Info("no new users for existing segment", "intent", intentName, "segment_id", seg.ID)
		return FindingOutcome{Action: outcomeSkippedNoNewUsers, IntentName: intentName, SegmentID: seg.ID}, nil
	}

	if err := uc.membership.BulkInsert(ctx, segUUID, newUsers); err != nil {
		return FindingOutcome{}, err
	}
	count, err := uc.membership.CountMembers(ctx, segUUID)
	if err != nil {
		return FindingOutcome{}, err
	}
	seg.ApproximateSize = count
	if err := uc.segments.Update(ctx, seg); err != nil {
		return FindingOutcome{}, err
	}

	uc.logger.Info("merged users into existing segment",
		"intent", intentName, "segment_id", seg.ID, "users_added", len(newUsers))
	return FindingOutcome{
		Action: outcomeMergedSegment, IntentName: intentName,
		UsersAdded: len(newUsers), SegmentID: seg.ID,
	}, nil
}

func (uc *ProcessIntentFindingUseCase) refreshPendingCandidate(
	ctx context.Context,
	id uuid.UUID,
	finding analyticsdomain.IntentConsistencyFinding,
	users []*domain.UserInCandidate,
) (FindingOutcome, error) {
	candidate := candidateFromFinding(finding)
	if err := uc.candidates.UpdateFromFinding(ctx, id, candidate, users); err != nil {
		return FindingOutcome{}, err
	}
	uc.logger.Info("refreshed pending segment candidate", "intent", finding.IntentName, "candidate_id", id)
	return FindingOutcome{
		Action: outcomeUpdatedCandidate, IntentName: finding.IntentName,
		CandidateID: id.String(), UsersAdded: len(users),
	}, nil
}

func (uc *ProcessIntentFindingUseCase) createCandidate(
	ctx context.Context,
	finding analyticsdomain.IntentConsistencyFinding,
	users []*domain.UserInCandidate,
) (FindingOutcome, error) {
	candidate := candidateFromFinding(finding)
	if err := uc.candidates.Save(ctx, candidate, users); err != nil {
		return FindingOutcome{}, err
	}
	uc.logger.Info("created segment candidate", "intent", finding.IntentName, "candidate_id", finding.FindingID)
	return FindingOutcome{
		Action: outcomeCreatedCandidate, IntentName: finding.IntentName,
		CandidateID: finding.FindingID.String(), UsersAdded: len(users),
	}, nil
}

func mapFindingUsers(finding analyticsdomain.IntentConsistencyFinding) []*domain.UserInCandidate {
	users := make([]*domain.UserInCandidate, 0, len(finding.Users))
	for _, u := range finding.Users {
		users = append(users, &domain.UserInCandidate{
			UserID: u.UserID, Confidence: u.Confidence,
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

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

func filterNewUsers(users []*domain.UserInCandidate, existing map[uuid.UUID]struct{}) []*domain.UserInCandidate {
	out := make([]*domain.UserInCandidate, 0, len(users))
	for _, u := range users {
		if _, ok := existing[u.UserID]; !ok {
			out = append(out, u)
		}
	}
	return out
}
