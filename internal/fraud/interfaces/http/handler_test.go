package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	fraudapp "skykin-platform/internal/fraud/application"
	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type stubSyncExecutor struct {
	execute func(context.Context, *time.Time) (*frauddomain.SyncSnapshot, error)
}

type stubReportExecutor struct {
	execute func(context.Context, fraudapp.IngestReportCommand) (*fraudapp.IngestReportResult, error)
}

func (s stubReportExecutor) Execute(
	ctx context.Context,
	command fraudapp.IngestReportCommand,
) (*fraudapp.IngestReportResult, error) {
	return s.execute(ctx, command)
}

func runReportRequest(
	t *testing.T,
	executor reportExecutor,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/reports", NewHandler(nil, executor).Report)
	request := httptest.NewRequest(http.MethodPost, "/reports", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func (s stubSyncExecutor) Execute(
	ctx context.Context,
	since *time.Time,
) (*frauddomain.SyncSnapshot, error) {
	return s.execute(ctx, since)
}

func runSyncRequest(t *testing.T, executor syncExecutor, target string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/sync", NewHandler(executor).Sync)
	request := httptest.NewRequest(http.MethodGet, target, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestSyncHandlerFullSnapshotWithoutCursor(t *testing.T) {
	cursor := time.Date(2026, 8, 3, 8, 30, 0, 123000000, time.UTC)
	executor := stubSyncExecutor{execute: func(_ context.Context, since *time.Time) (*frauddomain.SyncSnapshot, error) {
		require.Nil(t, since)
		return &frauddomain.SyncSnapshot{
			BlockedDomains: []frauddomain.BlockedDomain{},
			BlockedSenders: []frauddomain.BlockedSender{},
			ScamPatterns:   []frauddomain.ScamPattern{},
			NextCursor:     cursor,
		}, nil
	}}

	response := runSyncRequest(t, executor, "/sync")
	require.Equal(t, http.StatusOK, response.Code)

	var body SyncResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, "success", body.Status)
	require.Equal(t, "full", body.Mode)
	require.Equal(t, cursor.Format(time.RFC3339Nano), body.NextCursor)
	require.NotNil(t, body.BlockedDomains)
	require.NotNil(t, body.BlockedSenders)
	require.NotNil(t, body.ScamPatterns)
}

func TestSyncHandlerParsesDeltaCursor(t *testing.T) {
	since := time.Date(2026, 8, 3, 7, 0, 0, 0, time.UTC)
	executor := stubSyncExecutor{execute: func(_ context.Context, actual *time.Time) (*frauddomain.SyncSnapshot, error) {
		require.NotNil(t, actual)
		require.True(t, since.Equal(*actual))
		return &frauddomain.SyncSnapshot{
			BlockedDomains: []frauddomain.BlockedDomain{{
				Domain: "revoked.example", Status: frauddomain.StatusRevoked, UpdatedAt: since.Add(time.Minute),
			}},
			BlockedSenders: []frauddomain.BlockedSender{},
			ScamPatterns:   []frauddomain.ScamPattern{},
			NextCursor:     since.Add(time.Hour),
			IsDelta:        true,
		}, nil
	}}

	response := runSyncRequest(t, executor, "/sync?since="+since.Format(time.RFC3339))
	require.Equal(t, http.StatusOK, response.Code)

	var body SyncResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, "delta", body.Mode)
	require.Len(t, body.BlockedDomains, 1)
	require.Equal(t, frauddomain.StatusRevoked, body.BlockedDomains[0].Status)
}

func TestSyncHandlerRejectsInvalidCursor(t *testing.T) {
	called := false
	executor := stubSyncExecutor{execute: func(_ context.Context, _ *time.Time) (*frauddomain.SyncSnapshot, error) {
		called = true
		return nil, nil
	}}

	response := runSyncRequest(t, executor, "/sync?since=not-a-time")
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.False(t, called)
}

func TestSyncHandlerRejectsFutureCursor(t *testing.T) {
	executor := stubSyncExecutor{execute: func(_ context.Context, _ *time.Time) (*frauddomain.SyncSnapshot, error) {
		return nil, fraudapp.ErrFutureCursor
	}}

	response := runSyncRequest(t, executor, "/sync?since=2099-01-01T00:00:00Z")
	require.Equal(t, http.StatusBadRequest, response.Code)
}

func TestSyncHandlerMapsRepositoryFailure(t *testing.T) {
	executor := stubSyncExecutor{execute: func(_ context.Context, _ *time.Time) (*frauddomain.SyncSnapshot, error) {
		return nil, errors.New("database unavailable")
	}}

	response := runSyncRequest(t, executor, "/sync")
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.NotContains(t, response.Body.String(), "database unavailable")
}

func TestReportHandlerAcceptsOneReport(t *testing.T) {
	reportedAt := time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)
	executor := stubReportExecutor{execute: func(
		_ context.Context,
		command fraudapp.IngestReportCommand,
	) (*fraudapp.IngestReportResult, error) {
		require.Equal(t, "url_phishing", command.ThreatType)
		require.Equal(t, "https://scam.example/login", command.URLDomain)
		return &fraudapp.IngestReportResult{
			ReportID: "report-id", ReportedAt: reportedAt,
		}, nil
	}}
	response := runReportRequest(t, executor, `{
		"threat_type":"url_phishing",
		"severity":"high",
		"url_domain":"https://scam.example/login",
		"detection_source":"ml",
		"sdk_version":"1.0"
	}`)
	require.Equal(t, http.StatusAccepted, response.Code)

	var body ThreatReportAcceptedResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Equal(t, "accepted", body.Status)
	require.Equal(t, "report-id", body.ReportID)
	require.Equal(t, reportedAt.Format(time.RFC3339Nano), body.ReportedAt)
}

func TestReportHandlerMapsInvalidReportsToBadRequest(t *testing.T) {
	executor := stubReportExecutor{execute: func(
		_ context.Context,
		_ fraudapp.IngestReportCommand,
	) (*fraudapp.IngestReportResult, error) {
		return nil, fraudapp.ErrInvalidReport
	}}
	for _, body := range []string{
		`{"threat_type":"url_phishing","severity":"high","detection_source":"ml","sdk_version":"1.0"}`,
		`{"threat_type":"url_phishing","severity":"high","sender_hash":"BAD","detection_source":"ml","sdk_version":"1.0"}`,
		`{"threat_type":"url_phishing","severity":"high","url_domain":"not a host","detection_source":"ml","sdk_version":"1.0"}`,
	} {
		response := runReportRequest(t, executor, body)
		require.Equal(t, http.StatusBadRequest, response.Code)
	}
}

func TestReportHandlerMapsPersistenceAndQueueFailures(t *testing.T) {
	t.Run("persistence", func(t *testing.T) {
		executor := stubReportExecutor{execute: func(
			context.Context,
			fraudapp.IngestReportCommand,
		) (*fraudapp.IngestReportResult, error) {
			return nil, errors.New("database down")
		}}
		response := runReportRequest(t, executor, `{
			"threat_type":"url_phishing","severity":"high",
			"url_domain":"scam.example","detection_source":"ml","sdk_version":"1.0"
		}`)
		require.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("queue", func(t *testing.T) {
		executor := stubReportExecutor{execute: func(
			context.Context,
			fraudapp.IngestReportCommand,
		) (*fraudapp.IngestReportResult, error) {
			return &fraudapp.IngestReportResult{ReportID: "durable-id"}, fraudapp.ErrQueueUnavailable
		}}
		response := runReportRequest(t, executor, `{
			"threat_type":"url_phishing","severity":"high",
			"url_domain":"scam.example","detection_source":"ml","sdk_version":"1.0"
		}`)
		require.Equal(t, http.StatusServiceUnavailable, response.Code)
		require.Contains(t, response.Body.String(), "durable-id")
	})
}
