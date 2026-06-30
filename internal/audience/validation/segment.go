package validation

import (
	"errors"
	"strings"
)

// CreateSegmentInput is the validated shape for catalog segment creation.
type CreateSegmentInput struct {
	Name             string
	TopIntentSignals []string
	ApproximateSize  int
	EstimatedCPM     float64
}

// ValidateCreateSegment checks segment catalog fields before persistence.
func ValidateCreateSegment(in CreateSegmentInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errors.New("name is required")
	}
	if len(in.TopIntentSignals) == 0 {
		return errors.New("top_intent_signals must include at least one intent")
	}
	for _, sig := range in.TopIntentSignals {
		if strings.TrimSpace(sig) == "" {
			return errors.New("top_intent_signals cannot contain empty values")
		}
	}
	if in.ApproximateSize < 0 {
		return errors.New("approximate_size must be >= 0")
	}
	if in.EstimatedCPM <= 0 {
		return errors.New("estimated_cpm must be > 0")
	}
	return nil
}

// ValidateSegmentID ensures a segment id is present for lookups and updates.
func ValidateSegmentID(segmentID string) error {
	if strings.TrimSpace(segmentID) == "" {
		return errors.New("segment id is required")
	}
	return nil
}
