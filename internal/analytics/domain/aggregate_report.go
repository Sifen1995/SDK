package domain

import "time"

// AggregateReport is a batch of anonymized intent signals from one device.
type AggregateReport struct {
	DateBucket time.Time
	Intents    []AggregateIntentSignal
}

// AggregateIntentSignal is one intent rollup entry inside a device batch.
type AggregateIntentSignal struct {
	IntentName     string
	Count          int     // added to signal_count
	DaysConsistent float64 // added to weighted_count
}
