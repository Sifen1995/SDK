package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	adminApp "skykin-platform/internal/admin/application"
	adminEvents "skykin-platform/internal/admin/events"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	audienceApp "skykin-platform/internal/audience/application"
	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockIntentReader struct {
	byIntent map[string][]*analyticsdomain.ConsistentUser
	err      error
}

func (m *mockIntentReader) FindConsistentUsers(
	_ context.Context, intentName string, _ float64, _, _, _ int,
) ([]*analyticsdomain.ConsistentUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byIntent[intentName], nil
}

type mockCandidateRepo struct {
	byID    map[uuid.UUID]*audiencedomain.SegmentCandidate
	byIntent map[string]*audiencedomain.SegmentCandidate
	users   map[uuid.UUID][]*audiencedomain.UserInCandidate
	saved   []*audiencedomain.SegmentCandidate
	updated []uuid.UUID
	status  map[uuid.UUID]audiencedomain.CandidateStatus
	linked  map[uuid.UUID]uuid.UUID
}

func newMockCandidateRepo() *mockCandidateRepo {
	return &mockCandidateRepo{
		byID:     make(map[uuid.UUID]*audiencedomain.SegmentCandidate),
		byIntent: make(map[string]*audiencedomain.SegmentCandidate),
		users:    make(map[uuid.UUID][]*audiencedomain.UserInCandidate),
		status:   make(map[uuid.UUID]audiencedomain.CandidateStatus),
		linked:   make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *mockCandidateRepo) Save(_ context.Context, c *audiencedomain.SegmentCandidate, users []*audiencedomain.UserInCandidate) error {
	m.saved = append(m.saved, c)
	m.byID[c.ID] = c
	m.byIntent[c.IntentName] = c
	m.users[c.ID] = users
	return nil
}

func (m *mockCandidateRepo) FindByStatus(_ context.Context, status audiencedomain.CandidateStatus) ([]*audiencedomain.SegmentCandidate, error) {
	var out []*audiencedomain.SegmentCandidate
	for _, c := range m.byID {
		if c.Status == status {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *mockCandidateRepo) FindByID(_ context.Context, id uuid.UUID) (*audiencedomain.SegmentCandidate, error) {
	c, ok := m.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (m *mockCandidateRepo) FindPendingByIntentName(_ context.Context, intentName string) (*audiencedomain.SegmentCandidate, error) {
	c, ok := m.byIntent[intentName]
	if !ok || c.Status != audiencedomain.CandidateStatusPending {
		return nil, gorm.ErrRecordNotFound
	}
	return c, nil
}

func (m *mockCandidateRepo) UpdateFromFinding(_ context.Context, id uuid.UUID, c *audiencedomain.SegmentCandidate, users []*audiencedomain.UserInCandidate) error {
	m.updated = append(m.updated, id)
	m.byID[id] = c
	m.byIntent[c.IntentName] = c
	m.users[id] = users
	return nil
}

func (m *mockCandidateRepo) GetUsers(_ context.Context, candidateID uuid.UUID) ([]*audiencedomain.UserInCandidate, error) {
	return m.users[candidateID], nil
}

func (m *mockCandidateRepo) UpdateStatus(_ context.Context, id uuid.UUID, status audiencedomain.CandidateStatus, _ uuid.UUID, _ string) error {
	m.status[id] = status
	if c, ok := m.byID[id]; ok {
		c.Status = status
	}
	return nil
}

func (m *mockCandidateRepo) LinkToSegment(_ context.Context, candidateID, segmentID uuid.UUID) error {
	m.linked[candidateID] = segmentID
	return nil
}

type mockMembershipRepo struct {
	inserted map[string]struct{}
	segment  uuid.UUID
}

func newMockMembershipRepo() *mockMembershipRepo {
	return &mockMembershipRepo{inserted: make(map[string]struct{})}
}

func (m *mockMembershipRepo) BulkInsert(_ context.Context, segmentID uuid.UUID, users []*audiencedomain.UserInCandidate) error {
	m.segment = segmentID
	for _, u := range users {
		m.inserted[u.UserID] = struct{}{}
	}
	return nil
}

func (m *mockMembershipRepo) FindUsersInSegment(_ context.Context, _ uuid.UUID) ([]string, error) {
	out := make([]string, 0, len(m.inserted))
	for id := range m.inserted {
		out = append(out, id)
	}
	return out, nil
}

func (m *mockMembershipRepo) CountMembers(_ context.Context, _ uuid.UUID) (int, error) {
	return len(m.inserted), nil
}

type mockSegmentRepo struct {
	segments map[string]*audiencedomain.AudienceSegment
	updated  []*audiencedomain.AudienceSegment
}

func newMockSegmentRepo() *mockSegmentRepo {
	return &mockSegmentRepo{segments: make(map[string]*audiencedomain.AudienceSegment)}
}

func (m *mockSegmentRepo) GetByID(context.Context, string) (*audiencedomain.AudienceSegment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSegmentRepo) GetByName(context.Context, string) (*audiencedomain.AudienceSegment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSegmentRepo) Create(context.Context, *audiencedomain.AudienceSegment) error {
	return errors.New("not implemented")
}

func (m *mockSegmentRepo) Update(_ context.Context, seg *audiencedomain.AudienceSegment) error {
	m.updated = append(m.updated, seg)
	return nil
}

func (m *mockSegmentRepo) ListAvailableNow(_ context.Context, _ time.Time) ([]audiencedomain.AudienceSegment, error) {
	out := make([]audiencedomain.AudienceSegment, 0, len(m.segments))
	for _, seg := range m.segments {
		if seg.IsActive {
			out = append(out, *seg)
		}
	}
	return out, nil
}

func (m *mockSegmentRepo) ListAll(context.Context) ([]audiencedomain.AudienceSegment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSegmentRepo) FindActiveByIntentSignal(_ context.Context, intentName string, _ time.Time) (*audiencedomain.AudienceSegment, error) {
	seg, ok := m.segments[intentName]
	if !ok || !seg.IsActive {
		return nil, gorm.ErrRecordNotFound
	}
	return seg, nil
}

type mockProcessor struct {
	calls int
}

func (m *mockProcessor) Process(_ context.Context, finding analyticsdomain.IntentConsistencyFinding) (analyticsApp.FindingProcessResult, error) {
	m.calls++
	return analyticsApp.FindingProcessResult{
		Action: "created_candidate", IntentName: finding.IntentName, UsersAdded: finding.UserCount,
	}, nil
}

func TestAnalyzeIntentConsistency_ProcessesFindings(t *testing.T) {
	users := []*analyticsdomain.ConsistentUser{
		{UserID: uuid.New().String(), Confidence: 0.80, DaysActive: 6, LastSeenAt: time.Now()},
		{UserID: uuid.New().String(), Confidence: 0.90, DaysActive: 8, LastSeenAt: time.Now()},
	}
	processor := &mockProcessor{}
	cfg := analyticsdomain.ClassificationConfig{IntentClasses: []string{"coffee_interest"}}
	uc := analyticsApp.NewAnalyzeIntentConsistencyUseCase(
		&mockIntentReader{byIntent: map[string][]*analyticsdomain.ConsistentUser{"coffee_interest": users}},
		cfg, processor, nil,
	)
	report, err := uc.Run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, processor.calls)
	require.Equal(t, 1, report.CandidatesCreated)
}

func TestProcessIntentFinding_CreatesCandidateWhenNoMatch(t *testing.T) {
	repo := newMockCandidateRepo()
	segRepo := newMockSegmentRepo()
	memRepo := newMockMembershipRepo()
	uc := audienceApp.NewProcessIntentFindingUseCase(segRepo, memRepo, repo, nil)
	finding := analyticsdomain.IntentConsistencyFinding{
		FindingID: uuid.New(), IntentName: "crypto_interest", UserCount: 1,
		AvgConfidence: 0.9, AvgDaysActive: 6, MinDaysActive: 5, LookbackDays: 30,
		ScannedAt: time.Now().UTC(),
		Users: []*analyticsdomain.ConsistentUser{{UserID: uuid.New().String(), Confidence: 0.9, DaysActive: 6}},
	}
	outcome, err := uc.Execute(context.Background(), finding)
	require.NoError(t, err)
	assert.Equal(t, "created_candidate", outcome.Action)
	require.Len(t, repo.saved, 1)
}

func TestProcessIntentFinding_MergesIntoExistingSegment(t *testing.T) {
	repo := newMockCandidateRepo()
	segRepo := newMockSegmentRepo()
	memRepo := newMockMembershipRepo()
	segID := uuid.New()
	existingUser := uuid.New().String()
	newUser := uuid.New().String()
	memRepo.inserted[existingUser] = struct{}{}
	segRepo.segments["crypto_interest"] = &audiencedomain.AudienceSegment{
		ID: segID.String(), TopIntentSignals: []string{"crypto_interest"}, IsActive: true,
	}
	uc := audienceApp.NewProcessIntentFindingUseCase(segRepo, memRepo, repo, nil)
	finding := analyticsdomain.IntentConsistencyFinding{
		FindingID: uuid.New(), IntentName: "crypto_interest", UserCount: 1,
		AvgConfidence: 0.9, ScannedAt: time.Now().UTC(),
		Users: []*analyticsdomain.ConsistentUser{{UserID: newUser, Confidence: 0.9, DaysActive: 6}},
	}
	outcome, err := uc.Execute(context.Background(), finding)
	require.NoError(t, err)
	assert.Equal(t, "merged_segment", outcome.Action)
	assert.Equal(t, 1, outcome.UsersAdded)
	assert.Len(t, repo.saved, 0)
	_, ok := memRepo.inserted[newUser]
	assert.True(t, ok)
}

func TestProcessIntentFinding_SkipsWhenNoNewUsers(t *testing.T) {
	repo := newMockCandidateRepo()
	segRepo := newMockSegmentRepo()
	memRepo := newMockMembershipRepo()
	segID := uuid.New()
	existingUser := uuid.New().String()
	memRepo.inserted[existingUser] = struct{}{}
	segRepo.segments["crypto_interest"] = &audiencedomain.AudienceSegment{
		ID: segID.String(), TopIntentSignals: []string{"crypto_interest"}, IsActive: true,
	}
	uc := audienceApp.NewProcessIntentFindingUseCase(segRepo, memRepo, repo, nil)
	finding := analyticsdomain.IntentConsistencyFinding{
		FindingID: uuid.New(), IntentName: "crypto_interest", UserCount: 1,
		ScannedAt: time.Now().UTC(),
		Users:     []*analyticsdomain.ConsistentUser{{UserID: existingUser, Confidence: 0.9, DaysActive: 6}},
	}
	outcome, err := uc.Execute(context.Background(), finding)
	require.NoError(t, err)
	assert.Equal(t, "skipped_no_new_users", outcome.Action)
	assert.Len(t, repo.saved, 0)
}

func TestProcessIntentFinding_UpdatesPendingCandidate(t *testing.T) {
	repo := newMockCandidateRepo()
	pendingID := uuid.New()
	repo.byIntent["fashion_interest"] = &audiencedomain.SegmentCandidate{
		ID: pendingID, IntentName: "fashion_interest", Status: audiencedomain.CandidateStatusPending,
	}
	segRepo := newMockSegmentRepo()
	memRepo := newMockMembershipRepo()
	uc := audienceApp.NewProcessIntentFindingUseCase(segRepo, memRepo, repo, nil)
	finding := analyticsdomain.IntentConsistencyFinding{
		FindingID: uuid.New(), IntentName: "fashion_interest", UserCount: 2,
		AvgConfidence: 0.85, ScannedAt: time.Now().UTC(),
		Users: []*analyticsdomain.ConsistentUser{
			{UserID: uuid.New().String(), Confidence: 0.85, DaysActive: 6},
			{UserID: uuid.New().String(), Confidence: 0.85, DaysActive: 6},
		},
	}
	outcome, err := uc.Execute(context.Background(), finding)
	require.NoError(t, err)
	assert.Equal(t, "updated_candidate", outcome.Action)
	assert.Len(t, repo.saved, 0)
	assert.Len(t, repo.updated, 1)
}

func TestApproveCandidate_Success(t *testing.T) {
	candidateID := uuid.New()
	bus := messaging.NewBus()
	var published adminEvents.CandidateApprovedEvent
	done := make(chan struct{}, 1)
	bus.Subscribe(adminEvents.TopicCandidateApproved, func(e messaging.Event) {
		published = e.Payload.(adminEvents.CandidateApprovedEvent)
		done <- struct{}{}
	})
	uc := adminApp.NewApproveCandidateUseCase(bus, nil)

	err := uc.Execute(context.Background(), candidateID, uuid.New(), "Crypto Segment", "desc", 6.5)
	require.NoError(t, err)
	<-done
	assert.Equal(t, candidateID, published.CandidateID)
	assert.Equal(t, "Crypto Segment", published.Name)
	assert.Equal(t, 6.5, published.EstimatedCPM)
}

func TestRejectCandidate_Success(t *testing.T) {
	candidateID := uuid.New()
	bus := messaging.NewBus()
	var published adminEvents.CandidateRejectedEvent
	done := make(chan struct{}, 1)
	bus.Subscribe(adminEvents.TopicCandidateRejected, func(e messaging.Event) {
		published = e.Payload.(adminEvents.CandidateRejectedEvent)
		done <- struct{}{}
	})
	uc := adminApp.NewRejectCandidateUseCase(bus, nil)
	require.NoError(t, uc.Execute(context.Background(), candidateID, uuid.New(), "too small"))
	<-done
	assert.Equal(t, candidateID, published.CandidateID)
	assert.Equal(t, "too small", published.Notes)
}

func TestAudienceRejectCandidate_Success(t *testing.T) {
	candidateID := uuid.New()
	candidateRepo := newMockCandidateRepo()
	candidateRepo.byID[candidateID] = &audiencedomain.SegmentCandidate{ID: candidateID, Status: audiencedomain.CandidateStatusPending}
	uc := audienceApp.NewRejectCandidateUseCase(candidateRepo, nil)
	adminID := uuid.New()
	require.NoError(t, uc.Execute(context.Background(), adminEvents.CandidateRejectedEvent{
		CandidateID: candidateID,
		AdminID:     adminID,
		Notes:       "too small",
	}))
	assert.Equal(t, audiencedomain.CandidateStatusRejected, candidateRepo.status[candidateID])
}
