package analytics_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adminApp "skykin-platform/internal/admin/application"
	adminHTTP "skykin-platform/internal/admin/interfaces/http"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	audienceApp "skykin-platform/internal/audience/application"
	audiencedomain "skykin-platform/internal/audience/domain"
	audienceEvents "skykin-platform/internal/audience/interfaces/events"
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	"skykin-platform/internal/platform/messaging"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSegmentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	stmts := []string{
		`CREATE TABLE segment_candidates (
			id TEXT PRIMARY KEY, intent_name TEXT NOT NULL, user_count INTEGER NOT NULL DEFAULT 0,
			avg_confidence REAL NOT NULL, avg_days_active REAL NOT NULL, min_days_active INTEGER NOT NULL,
			lookback_days INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
			scanned_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			reviewed_by TEXT, reviewed_at DATETIME, review_notes TEXT, published_segment_id TEXT
		)`,
		`CREATE TABLE audience_segments (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT,
			top_intent_signals TEXT NOT NULL DEFAULT '[]', approximate_size INTEGER NOT NULL DEFAULT 0,
			estimated_cpm REAL NOT NULL, available_from DATETIME NOT NULL,
			available_until DATETIME, is_active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE segment_memberships (
			segment_id TEXT NOT NULL, user_id TEXT NOT NULL,
			confidence REAL NOT NULL, days_active INTEGER NOT NULL,
			added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (segment_id, user_id)
		)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
	return db
}

func seedPendingCandidate(t *testing.T, repo *audienceInfra.CandidateRepository) uuid.UUID {
	t.Helper()
	candidateID := uuid.New()
	candidate := &audiencedomain.SegmentCandidate{
		ID: candidateID, IntentName: "fashion_interest", UserCount: 1,
		AvgConfidence: 0.82, AvgDaysActive: 6, MinDaysActive: 5, LookbackDays: 30,
		Status: audiencedomain.CandidateStatusPending, ScannedAt: time.Now().UTC(),
	}
	users := []*audiencedomain.UserInCandidate{
		{UserID: uuid.New().String(), Confidence: 0.82, DaysActive: 6, LastSeenAt: time.Now().UTC()},
	}
	require.NoError(t, repo.Save(context.Background(), candidate, users))
	return candidateID
}

func TestAudience_ListSegmentCandidates_HTTP(t *testing.T) {
	db := setupSegmentTestDB(t)
	repo := audienceInfra.NewCandidateRepository(db)
	candidateID := seedPendingCandidate(t, repo)
	handler := audienceHTTP.NewHandler(nil, audienceApp.NewListSegmentCandidatesUseCase(repo))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/ad-portal/admin/audience/segment-candidates", handler.ListSegmentCandidates)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ad-portal/admin/audience/segment-candidates?status=pending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body []audienceHTTP.CandidateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, candidateID.String(), body[0].ID)
}

func TestAdmin_TriggerIntentConsistency_HTTP(t *testing.T) {
	processor := &mockProcessor{}
	uc := analyticsApp.NewAnalyzeIntentConsistencyUseCase(
		&noopIntentReader{}, analyticsdomain.ClassificationConfig{}, processor, slog.Default(),
	)
	handler := adminHTTP.NewAnalyticsHandler(uc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/ad-portal/admin/analytics/intent-consistency/run", handler.TriggerIntentConsistency)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ad-portal/admin/analytics/intent-consistency/run", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body analyticsApp.RunReport
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "Scan complete. No new segment candidates.", body.Message)
}

type noopIntentReader struct{}

func (noopIntentReader) FindConsistentUsers(context.Context, string, float64, int, int, int) ([]*analyticsdomain.ConsistentUser, error) {
	return nil, nil
}

type mockProcessor struct {
	calls int
}

func (m *mockProcessor) Process(_ context.Context, _ analyticsdomain.IntentConsistencyFinding) (analyticsApp.FindingProcessResult, error) {
	m.calls++
	return analyticsApp.FindingProcessResult{}, nil
}

func TestAdmin_ApproveSegmentCandidate_HTTP(t *testing.T) {
	db := setupSegmentTestDB(t)
	bus := messaging.NewBus()
	candidateRepo := audienceInfra.NewCandidateRepository(db)
	segmentRepo := audienceInfra.NewSegmentRepository(db)
	membershipRepo := audienceInfra.NewMembershipRepository(db)
	candidateID := seedPendingCandidate(t, candidateRepo)

	processApproved := audienceApp.NewProcessApprovedCandidateUseCase(
		segmentRepo, membershipRepo, candidateRepo, slog.Default(),
	)
	rejectCandidate := audienceApp.NewRejectCandidateUseCase(candidateRepo, slog.Default())
	recordPurchase := audienceApp.NewRecordSegmentPurchaseUseCase(audienceInfra.NewPurchaseRepository(db))
	audienceEvents.NewCandidateConsumer(processApproved, rejectCandidate, recordPurchase, slog.Default()).Register(bus)

	approveUC := adminApp.NewApproveCandidateUseCase(bus, slog.Default())
	handler := adminHTTP.NewSegmentCandidateHandler(approveUC, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	adminID := uuid.New()
	r.Use(func(c *gin.Context) { c.Set("portal_user_id", adminID.String()); c.Next() })
	r.POST("/api/v1/ad-portal/admin/audience/segment-candidates/:id/approve", handler.ApproveSegmentCandidate)

	payload := `{"name":"Fashion","description":"desc","estimated_cpm":4.25}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ad-portal/admin/audience/segment-candidates/"+candidateID.String()+"/approve", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code, w.Body.String())

	require.Eventually(t, func() bool {
		candidate, err := candidateRepo.FindByID(context.Background(), candidateID)
		if err != nil || candidate.PublishedSegmentID == nil {
			return false
		}
		return true
	}, time.Second, 10*time.Millisecond)
}
