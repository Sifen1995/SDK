package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	adminApp "skykin-platform/internal/admin/application"
	analyticsApp "skykin-platform/internal/analytics/application"
	analyticsdomain "skykin-platform/internal/analytics/domain"
	audienceApp "skykin-platform/internal/audience/application"
	audiencedomain "skykin-platform/internal/audience/domain"
	"skykin-platform/internal/audience/model"
	"skykin-platform/internal/platform/messaging"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	byID   map[uuid.UUID]*audiencedomain.SegmentCandidate
	users  map[uuid.UUID][]*audiencedomain.UserInCandidate
	saved  []*audiencedomain.SegmentCandidate
	status map[uuid.UUID]audiencedomain.CandidateStatus
	linked map[uuid.UUID]uuid.UUID
}

func newMockCandidateRepo() *mockCandidateRepo {
	return &mockCandidateRepo{
		byID: make(map[uuid.UUID]*audiencedomain.SegmentCandidate),
		users: make(map[uuid.UUID][]*audiencedomain.UserInCandidate),
		status: make(map[uuid.UUID]audiencedomain.CandidateStatus),
		linked: make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *mockCandidateRepo) Save(_ context.Context, c *audiencedomain.SegmentCandidate, users []*audiencedomain.UserInCandidate) error {
	m.saved = append(m.saved, c)
	m.byID[c.ID] = c
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
	inserted []uuid.UUID
	segment  uuid.UUID
}

func (m *mockMembershipRepo) BulkInsert(_ context.Context, segmentID uuid.UUID, users []*audiencedomain.UserInCandidate) error {
	m.segment = segmentID
	for _, u := range users {
		m.inserted = append(m.inserted, u.UserID)
	}
	return nil
}

func (m *mockMembershipRepo) FindUsersInSegment(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return m.inserted, nil
}

type mockCatalog struct {
	segment *model.AudienceSegment
}

func (m *mockCatalog) CreateSegment(_ context.Context, cmd adminApp.CreateSegmentCmd) (*model.AudienceSegment, error) {
	return m.segment, nil
}

func TestAnalyzeIntentConsistency_PublishesFindings(t *testing.T) {
	users := []*analyticsdomain.ConsistentUser{
		{UserID: uuid.New(), Confidence: 0.80, DaysActive: 6, LastSeenAt: time.Now()},
		{UserID: uuid.New(), Confidence: 0.90, DaysActive: 8, LastSeenAt: time.Now()},
	}
	var received []analyticsdomain.IntentConsistencyFinding
	bus := messaging.NewBus()
	done := make(chan struct{}, 1)
	bus.Subscribe(analyticsdomain.TopicIntentConsistencyFinding, func(e messaging.Event) {
		received = append(received, e.Payload.(analyticsdomain.IntentConsistencyFinding))
		done <- struct{}{}
	})
	cfg := analyticsdomain.ClassificationConfig{IntentClasses: []string{"coffee_interest"}}
	uc := analyticsApp.NewAnalyzeIntentConsistencyUseCase(
		&mockIntentReader{byIntent: map[string][]*analyticsdomain.ConsistentUser{"coffee_interest": users}},
		cfg, bus, nil,
	)
	require.NoError(t, uc.Run(context.Background()))
	<-done
	require.Len(t, received, 1)
	assert.Equal(t, "coffee_interest", received[0].IntentName)
	assert.Equal(t, 2, received[0].UserCount)
}

func TestSaveCandidateFromFinding(t *testing.T) {
	repo := newMockCandidateRepo()
	uc := audienceApp.NewSaveCandidateFromFindingUseCase(repo, nil)
	finding := analyticsdomain.IntentConsistencyFinding{
		FindingID: uuid.New(), IntentName: "crypto_interest", UserCount: 1,
		AvgConfidence: 0.9, AvgDaysActive: 6, MinDaysActive: 5, LookbackDays: 30,
		ScannedAt: time.Now().UTC(),
		Users: []*analyticsdomain.ConsistentUser{{UserID: uuid.New(), Confidence: 0.9, DaysActive: 6}},
	}
	require.NoError(t, uc.Execute(context.Background(), finding))
	require.Len(t, repo.saved, 1)
	assert.Equal(t, audiencedomain.CandidateStatusPending, repo.saved[0].Status)
}

func TestApproveCandidate_Success(t *testing.T) {
	candidateID := uuid.New()
	segmentID := uuid.New()
	userID := uuid.New()
	candidateRepo := newMockCandidateRepo()
	candidateRepo.byID[candidateID] = &audiencedomain.SegmentCandidate{
		ID: candidateID, IntentName: "crypto_interest", UserCount: 1,
		Status: audiencedomain.CandidateStatusPending,
	}
	candidateRepo.users[candidateID] = []*audiencedomain.UserInCandidate{{UserID: userID, Confidence: 0.88, DaysActive: 7}}
	membershipRepo := &mockMembershipRepo{}
	catalog := &mockCatalog{segment: &model.AudienceSegment{ID: segmentID.String(), Name: "Crypto Segment"}}
	uc := adminApp.NewApproveCandidateUseCase(candidateRepo, membershipRepo, catalog, nil)

	seg, err := uc.Execute(context.Background(), candidateID, uuid.New(), "Crypto Segment", "desc", 6.5)
	require.NoError(t, err)
	require.NotNil(t, seg)
	assert.Equal(t, audiencedomain.CandidateStatusApproved, candidateRepo.status[candidateID])
	assert.Equal(t, []uuid.UUID{userID}, membershipRepo.inserted)
}

func TestRejectCandidate_Success(t *testing.T) {
	candidateID := uuid.New()
	candidateRepo := newMockCandidateRepo()
	candidateRepo.byID[candidateID] = &audiencedomain.SegmentCandidate{ID: candidateID, Status: audiencedomain.CandidateStatusPending}
	uc := adminApp.NewRejectCandidateUseCase(candidateRepo, nil)
	require.NoError(t, uc.Execute(context.Background(), candidateID, uuid.New(), "too small"))
	assert.Equal(t, audiencedomain.CandidateStatusRejected, candidateRepo.status[candidateID])
}
