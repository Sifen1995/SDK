package persistence

import "time"

// IntentAggregateCountRow is the GORM model for intent_aggregate_counts.
// Analytics owns this anonymous rollup table.
type IntentAggregateCountRow struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IntentName    string    `gorm:"type:varchar(100);not null;uniqueIndex:uq_intent_date"`
	DateBucket    time.Time `gorm:"type:date;not null;default:CURRENT_DATE;uniqueIndex:uq_intent_date"`
	SignalCount   int       `gorm:"not null;default:0"`
	WeightedCount float64   `gorm:"type:numeric(10,2);not null;default:0"`
}

func (IntentAggregateCountRow) TableName() string { return "intent_aggregate_counts" }
