package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"
	platformredis "skykin-platform/internal/platform/redis"
)

const ThreatReportQueueKey = "queue:fraud_threat_reports"

type ThreatReportQueuePayload struct {
	ID              string  `json:"id"`
	ThreatType      string  `json:"threat_type"`
	Severity        string  `json:"severity"`
	SenderHash      *string `json:"sender_hash,omitempty"`
	URLDomain       *string `json:"url_domain,omitempty"`
	DetectionSource string  `json:"detection_source"`
	SDKVersion      string  `json:"sdk_version"`
	ReportedAt      string  `json:"reported_at"`
}

type ThreatReportQueue struct {
	rdb *platformredis.RedisClient
}

func NewThreatReportQueue(rdb *platformredis.RedisClient) *ThreatReportQueue {
	return &ThreatReportQueue{rdb: rdb}
}

func (q *ThreatReportQueue) Enqueue(
	ctx context.Context,
	report *frauddomain.ThreatReport,
) error {
	if q == nil || q.rdb == nil {
		return fmt.Errorf("threat report queue is not configured")
	}
	if report == nil {
		return fmt.Errorf("threat report is required")
	}
	raw, err := encodeThreatReportQueuePayload(report)
	if err != nil {
		return err
	}
	return q.rdb.RPush(ctx, ThreatReportQueueKey, raw)
}

func encodeThreatReportQueuePayload(report *frauddomain.ThreatReport) (string, error) {
	raw, err := json.Marshal(ThreatReportQueuePayload{
		ID:              report.ID,
		ThreatType:      report.ThreatType,
		Severity:        report.Severity,
		SenderHash:      report.SenderHash,
		URLDomain:       report.URLDomain,
		DetectionSource: report.DetectionSource,
		SDKVersion:      report.SDKVersion,
		ReportedAt:      report.ReportedAt.UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func DecodeThreatReportQueuePayload(raw string) (*frauddomain.ThreatReport, error) {
	var payload ThreatReportQueuePayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}
	reportedAt, err := time.Parse(time.RFC3339Nano, payload.ReportedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid reported_at: %w", err)
	}
	if payload.ID == "" {
		return nil, fmt.Errorf("report id is required")
	}
	return &frauddomain.ThreatReport{
		ID:              payload.ID,
		ThreatType:      payload.ThreatType,
		Severity:        payload.Severity,
		SenderHash:      payload.SenderHash,
		URLDomain:       payload.URLDomain,
		DetectionSource: payload.DetectionSource,
		SDKVersion:      payload.SDKVersion,
		ReportedAt:      reportedAt.UTC(),
	}, nil
}
