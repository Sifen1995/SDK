package application

import (
	"context"

	"skykin-platform/internal/analytics/domain"
)

// IntentFindingProcessor handles a single intent consistency finding (implemented by audience module).
type IntentFindingProcessor interface {
	Process(ctx context.Context, finding domain.IntentConsistencyFinding) (FindingProcessResult, error)
}

// FindingProcessResult is the cross-module outcome of processing one finding.
type FindingProcessResult struct {
	Action      string `json:"action"`
	IntentName  string `json:"intent_name"`
	UsersAdded  int    `json:"users_added"`
	SegmentID   string `json:"segment_id,omitempty"`
	CandidateID string `json:"candidate_id,omitempty"`
}

// RunReport summarizes a full intent consistency scan.
type RunReport struct {
	CandidatesCreated int      `json:"candidates_created"`
	CandidatesUpdated int      `json:"candidates_updated"`
	SegmentsEnriched  int      `json:"segments_enriched"`
	UsersAdded        int      `json:"users_added_to_segments"`
	IntentsSkipped    []string `json:"intents_skipped"`
	Message           string   `json:"message"`
}
