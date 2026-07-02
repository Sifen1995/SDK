package application

import (
	"context"

	"skykin-platform/internal/audience/domain"
)

type ListSegmentCandidatesUseCase struct {
	repo domain.CandidateRepository
}

func NewListSegmentCandidatesUseCase(repo domain.CandidateRepository) *ListSegmentCandidatesUseCase {
	return &ListSegmentCandidatesUseCase{repo: repo}
}

func (uc *ListSegmentCandidatesUseCase) ListByStatus(ctx context.Context, status domain.CandidateStatus) ([]*domain.SegmentCandidate, error) {
	return uc.repo.FindByStatus(ctx, status)
}
