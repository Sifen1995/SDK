package application

import (
	"fmt"

	"skykin-platform/internal/audience/domain"
)

// ParseIntentSignals returns top_intent_signals from a segment definition.
func ParseIntentSignals(seg *domain.AudienceSegment) ([]string, error) {
	if seg == nil || len(seg.TopIntentSignals) == 0 {
		return nil, fmt.Errorf("segment has no intent signals")
	}
	return seg.TopIntentSignals, nil
}

// TargetIntentAllowed returns true when targetIntent appears in the segment signal list.
func TargetIntentAllowed(signals []string, targetIntent string) bool {
	for _, s := range signals {
		if s == targetIntent {
			return true
		}
	}
	return false
}
