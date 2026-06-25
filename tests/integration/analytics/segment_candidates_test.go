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
	audienceHTTP "skykin-platform/internal/audience/interfaces/http"
	audienceInfra "skykin-platform/internal/audience/infrastructure"
	"skykin-platform/internal/audience/model"
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
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	stmts := []string{
		`CREATE TABLE segment_candidates (
			id TEXT PRIMARY KEY, intent_name TEXT NOT NULL, user_count INTEGER NOT NULL DEFAULT 0,
			avg_confidence REAL NOT NULL, avg_days_active REAL NOT NULL, min_days_active INTEGER NOT NULL,
			lookback_days INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
			scanned_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			reviewed_by TEXT, reviewed_at DATETIME, review_notes TEXT, published_segment_id TEXT
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
		{UserID: uuid.New(), Confidence: 0.82, DaysActive: 6, LastSeenAt: time.Now().UTC()},
	}
	require.NoError(t, repo.Save(context.Background(), candidate, users))
	return candidateID
}

type stubPublisher struct {
	segmentID uuid.UUID
}

func (s *stubPublisher) CreateSegment(_ context.Context, cmd adminApp.CreateSegmentCmd) (*model.AudienceSegment, error) {
	return &model.AudienceSegment{ID: s.segmentID.String(), Name: cmd.Name}, nil
}

func TestAudience_ListSegmentCandidates_HTTP(t *testing.T) {
	db := setupSegmentTestDB(t)
	repo := audienceInfra.NewCandidateRepository(db)
	candidateID := seedPendingCandidate(t, repo)
	handler := audienceHTTP.NewHandler(nil, audienceApp.NewListSegmentCandidatesUseCase(repo))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/ad-portal/audience/segment-candidates", handler.ListSegmentCandidates)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ad-portal/audience/segment-candidates?status=pending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body []audienceHTTP.CandidateResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, candidateID.String(), body[0].ID)
}

func TestAdmin_TriggerIntentConsistency_HTTP(t *testing.T) {
	bus := messaging.NewBus()
	uc := analyticsApp.NewAnalyzeIntentConsistencyUseCase(&noopIntentReader{}, analyticsdomain.ClassificationConfig{}, bus, slog.Default())
	handler := adminHTTP.NewAnalyticsHandler(uc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/ad-portal/admin/analytics/intent-consistency/run", handler.TriggerIntentConsistency)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ad-portal/admin/analytics/intent-consistency/run", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusAccepted, w.Code)
}

func TestAdmin_ApproveSegmentCandidate_HTTP(t *testing.T) {
	candidateID := uuid.New()
	segmentID := uuid.New()
	adminID := uuid.New()
	repo := newInMemoryCandidateRepo()
	repo.byID[candidateID] = &audiencedomain.SegmentCandidate{
		ID: candidateID, IntentName: "fashion_interest", UserCount: 1, Status: audiencedomain.CandidateStatusPending,
	}
	repo.users[candidateID] = []*audiencedomain.UserInCandidate{{UserID: uuid.New(), Confidence: 0.82, DaysActive: 6}}
	handler := adminHTTP.NewSegmentCandidateHandler(
		adminApp.NewApproveCandidateUseCase(repo, &recordingMembershipRepo{}, &stubPublisher{segmentID: segmentID}, nil),
		nil,
	)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("portal_user_id", adminID.String()); c.Next() })
	r.POST("/api/v1/ad-portal/admin/audience/segment-candidates/:id/approve", handler.ApproveSegmentCandidate)

	payload := `{"name":"Fashion","description":"desc","estimated_cpm":4.25}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ad-portal/admin/audience/segment-candidates/"+candidateID.String()+"/approve", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
}

type noopIntentReader struct{}

func (noopIntentReader) FindConsistentUsers(context.Context, string, float64, int, int, int) ([]*analyticsdomain.ConsistentUser, error) {
	return nil, nil
}

type inMemoryCandidateRepo struct {
	byID   map[uuid.UUID]*audiencedomain.SegmentCandidate
	users  map[uuid.UUID][]*audiencedomain.UserInCandidate
	linked map[uuid.UUID]uuid.UUID
}

func newInMemoryCandidateRepo() *inMemoryCandidateRepo {
	return &inMemoryCandidateRepo{
		byID: make(map[uuid.UUID]*audiencedomain.SegmentCandidate),
		users: make(map[uuid.UUID][]*audiencedomain.UserInCandidate),
		linked: make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *inMemoryCandidateRepo) Save(_ context.Context, c *audiencedomain.SegmentCandidate, users []*audiencedomain.UserInCandidate) error {
	m.byID[c.ID] = c
	m.users[c.ID] = users
	return nil
}
func (m *inMemoryCandidateRepo) FindByStatus(context.Context, audiencedomain.CandidateStatus) ([]*audiencedomain.SegmentCandidate, error) {
	return nil, nil
}
func (m *inMemoryCandidateRepo) FindByID(_ context.Context, id uuid.UUID) (*audiencedomain.SegmentCandidate, error) {
	return m.byID[id], nil
}
func (m *inMemoryCandidateRepo) GetUsers(_ context.Context, id uuid.UUID) ([]*audiencedomain.UserInCandidate, error) {
	return m.users[id], nil
}
func (m *inMemoryCandidateRepo) UpdateStatus(_ context.Context, id uuid.UUID, status audiencedomain.CandidateStatus, _ uuid.UUID, _ string) error {
	m.byID[id].Status = status
	return nil
}
func (m *inMemoryCandidateRepo) LinkToSegment(_ context.Context, candidateID, segmentID uuid.UUID) error {
	m.linked[candidateID] = segmentID
	return nil
}

type recordingMembershipRepo struct{}

func (recordingMembershipRepo) BulkInsert(context.Context, uuid.UUID, []*audiencedomain.UserInCandidate) error {
	return nil
}
func (recordingMembershipRepo) FindUsersInSegment(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
