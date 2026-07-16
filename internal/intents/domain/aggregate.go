package domain

import "time"

// IntentAggregateCount is a daily rollup of anonymized intent signals.
type IntentAggregateCount struct {
	ID            string
	IntentName    string
	DateBucket    time.Time // date only (UTC midnight)
	SignalCount   int
	WeightedCount float64
}

// IntentAggregateSignal is an incoming aggregate signal to apply for a date bucket.
type IntentAggregateSignal struct {
	IntentName     string
	Count          int     // added to signal_count (defaults to 1 when unset)
	DaysConsistent float64 // contributes to weighted_count (7-day > 5-day)
	DateBucket     time.Time // zero = today UTC
}
