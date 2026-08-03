package application

import (
	"context"
	"errors"
	"testing"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/stretchr/testify/require"
)

type stubReportRepository struct {
	create func(*frauddomain.ThreatReport) error
}

func (s stubReportRepository) Create(_ context.Context, report *frauddomain.ThreatReport) error {
	return s.create(report)
}
func (stubReportRepository) HighestSeverityForDomain(context.Context, string, time.Time) (string, error) {
	return "", nil
}
func (stubReportRepository) HighestSeverityForSender(context.Context, string, time.Time) (string, error) {
	return "", nil
}

type stubReportQueue struct {
	enqueue func(*frauddomain.ThreatReport) error
}

func (s stubReportQueue) Enqueue(_ context.Context, report *frauddomain.ThreatReport) error {
	return s.enqueue(report)
}

func validReportCommand() IngestReportCommand {
	return IngestReportCommand{
		ThreatType:      "url_phishing",
		Severity:        "high",
		URLDomain:       "HTTPS://Example.COM:443/login?source=sms#verify",
		DetectionSource: "ml",
		SDKVersion:      "1.2.0",
	}
}

func TestIngestReportNormalizesAndPersistsBeforeEnqueue(t *testing.T) {
	now := time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)
	persisted := false
	repository := stubReportRepository{create: func(report *frauddomain.ThreatReport) error {
		persisted = true
		require.Equal(t, "example.com", *report.URLDomain)
		require.Equal(t, "report-id", report.ID)
		require.Equal(t, now, report.ReportedAt)
		return nil
	}}
	queue := stubReportQueue{enqueue: func(report *frauddomain.ThreatReport) error {
		require.True(t, persisted)
		require.Equal(t, "example.com", *report.URLDomain)
		return nil
	}}
	useCase := NewIngestReportUseCase(repository, queue)
	useCase.now = func() time.Time { return now }
	useCase.newID = func() string { return "report-id" }

	result, err := useCase.Execute(context.Background(), validReportCommand())
	require.NoError(t, err)
	require.Equal(t, "report-id", result.ReportID)
	require.Equal(t, now, result.ReportedAt)
}

func TestIngestReportRejectsInvalidIndicatorsBeforePersistence(t *testing.T) {
	called := false
	repository := stubReportRepository{create: func(*frauddomain.ThreatReport) error {
		called = true
		return nil
	}}
	useCase := NewIngestReportUseCase(repository, nil)

	tests := []IngestReportCommand{
		{
			ThreatType: "url_phishing", Severity: "high",
			DetectionSource: "ml", SDKVersion: "1.0",
		},
		{
			ThreatType: "url_phishing", Severity: "high",
			SenderHash: "ABC123", DetectionSource: "ml", SDKVersion: "1.0",
		},
		{
			ThreatType: "url_phishing", Severity: "high",
			URLDomain: "https://user:pass@example.com", DetectionSource: "ml", SDKVersion: "1.0",
		},
	}
	for _, command := range tests {
		_, err := useCase.Execute(context.Background(), command)
		require.ErrorIs(t, err, ErrInvalidReport)
	}
	require.False(t, called)
}

func TestIngestReportReturnsQueueErrorAfterDurableInsert(t *testing.T) {
	repository := stubReportRepository{create: func(*frauddomain.ThreatReport) error {
		return nil
	}}
	queue := stubReportQueue{enqueue: func(*frauddomain.ThreatReport) error {
		return errors.New("redis down")
	}}
	useCase := NewIngestReportUseCase(repository, queue)
	useCase.newID = func() string { return "durable-report" }

	result, err := useCase.Execute(context.Background(), validReportCommand())
	require.ErrorIs(t, err, ErrQueueUnavailable)
	require.Equal(t, "durable-report", result.ReportID)
}
