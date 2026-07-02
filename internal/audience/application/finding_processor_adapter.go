package application

import (
	"context"

	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsdomain "skykin-platform/internal/analytics/domain"
)

// FindingProcessorAdapter maps audience outcomes to analytics processor results.
type FindingProcessorAdapter struct {
	inner *ProcessIntentFindingUseCase
}

func NewFindingProcessorAdapter(inner *ProcessIntentFindingUseCase) *FindingProcessorAdapter {
	return &FindingProcessorAdapter{inner: inner}
}

func (a *FindingProcessorAdapter) Process(
	ctx context.Context,
	finding analyticsdomain.IntentConsistencyFinding,
) (analyticsApp.FindingProcessResult, error) {
	outcome, err := a.inner.Execute(ctx, finding)
	if err != nil {
		return analyticsApp.FindingProcessResult{}, err
	}
	return analyticsApp.FindingProcessResult{
		Action: outcome.Action, IntentName: outcome.IntentName,
		UsersAdded: outcome.UsersAdded, SegmentID: outcome.SegmentID,
		CandidateID: outcome.CandidateID,
	}, nil
}

var _ analyticsApp.IntentFindingProcessor = (*FindingProcessorAdapter)(nil)
