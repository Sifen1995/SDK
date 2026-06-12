package application

import (
	"encoding/json"
	"fmt"

	"skykin-platform/internal/audience/model"
)

// ParseIntentSignals unmarshals top_intent_signals JSON into intent slug strings.
func ParseIntentSignals(seg *model.AudienceSegment) ([]string, error) {
	if seg == nil || len(seg.TopIntentSignals) == 0 {
		return nil, fmt.Errorf("segment has no intent signals")
	}
	var signals []string
	if err := json.Unmarshal(seg.TopIntentSignals, &signals); err != nil {
		return nil, fmt.Errorf("invalid top_intent_signals: %w", err)
	}
	if len(signals) == 0 {
		return nil, fmt.Errorf("segment has no intent signals")
	}
	return signals, nil
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
