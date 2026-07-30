package bootstrap

import (
	"context"
	"errors"
	"log/slog"

	adminApp "skykin-platform/internal/admin/application"
	audienceApp "skykin-platform/internal/audience/application"
	audienceInfra "skykin-platform/internal/audience/infrastructure"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SegmentReviewPorts adapts audience use cases to the ports the admin module owns.
type SegmentReviewPorts struct {
	Publisher adminApp.SegmentPublisher
	Rejecter  adminApp.CandidateRejecter
}

// NewSegmentReviewPorts wires the transactional audience review use cases.
func NewSegmentReviewPorts(db *gorm.DB, logger *slog.Logger) *SegmentReviewPorts {
	if logger == nil {
		logger = slog.Default()
	}
	uow := audienceInfra.NewUnitOfWork(db)
	return &SegmentReviewPorts{
		Publisher: &segmentPublisherAdapter{approve: audienceApp.NewProcessApprovedCandidateUseCase(uow, logger)},
		Rejecter:  &candidateRejecterAdapter{reject: audienceApp.NewRejectCandidateUseCase(uow, logger)},
	}
}

type segmentPublisherAdapter struct {
	approve *audienceApp.ProcessApprovedCandidateUseCase
}

func (a *segmentPublisherAdapter) PublishFromCandidate(
	ctx context.Context,
	candidateID, adminID uuid.UUID,
	name, description string,
	estimatedCPM float64,
) (adminApp.ApprovedSegment, error) {
	published, err := a.approve.Execute(ctx, audienceApp.ApproveCandidateCmd{
		CandidateID:  candidateID,
		AdminID:      adminID,
		Name:         name,
		Description:  description,
		EstimatedCPM: estimatedCPM,
	})
	if err != nil {
		return adminApp.ApprovedSegment{}, translateCandidateError(err)
	}
	return adminApp.ApprovedSegment{
		SegmentID:   published.SegmentID,
		MemberCount: published.MemberCount,
	}, nil
}

type candidateRejecterAdapter struct {
	reject *audienceApp.RejectCandidateUseCase
}

func (a *candidateRejecterAdapter) RejectCandidate(ctx context.Context, candidateID, adminID uuid.UUID, notes string) error {
	return translateCandidateError(a.reject.Execute(ctx, candidateID, adminID, notes))
}

func translateCandidateError(err error) error {
	if errors.Is(err, audienceApp.ErrCandidateNotPending) {
		return adminApp.ErrCandidateNotPending
	}
	return err
}
