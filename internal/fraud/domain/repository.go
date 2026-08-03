package domain

import (
	"context"
	"time"
)

// SyncRepository reads a stable full snapshot or timestamp-bounded delta.
type SyncRepository interface {
	Sync(ctx context.Context, since *time.Time, until time.Time) (*SyncSnapshot, error)
}

type ThreatReportRepository interface {
	Create(ctx context.Context, report *ThreatReport) error
	HighestSeverityForDomain(ctx context.Context, domain string, since time.Time) (string, error)
	HighestSeverityForSender(ctx context.Context, senderHash string, since time.Time) (string, error)
}

type PromotionRepository interface {
	PromoteDomain(
		ctx context.Context,
		domain, threatType, severity string,
		now time.Time,
	) error
	PromoteSender(
		ctx context.Context,
		senderHash, threatType, severity string,
		now time.Time,
	) error
}
