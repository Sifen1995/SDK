package infrastructure

import (
	"context"
	"strings"
	"testing"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/stretchr/testify/require"
)

type fakeReportRedis struct {
	sets map[string]map[string]float64
}

func newFakeReportRedis() *fakeReportRedis {
	return &fakeReportRedis{
		sets: make(map[string]map[string]float64),
	}
}

func (*fakeReportRedis) BRPop(context.Context, string, time.Duration) (string, error) {
	return "", nil
}
func (f *fakeReportRedis) ZAdd(_ context.Context, key, member string, score float64) error {
	if f.sets[key] == nil {
		f.sets[key] = make(map[string]float64)
	}
	f.sets[key][member] = score
	return nil
}
func (f *fakeReportRedis) ZRemRangeByScore(
	_ context.Context,
	key string,
	min, max float64,
) error {
	for member, score := range f.sets[key] {
		if score >= min && score <= max {
			delete(f.sets[key], member)
		}
	}
	return nil
}
func (f *fakeReportRedis) ZCard(_ context.Context, key string) (int64, error) {
	return int64(len(f.sets[key])), nil
}
func (*fakeReportRedis) Expire(context.Context, string, time.Duration) error { return nil }

type fakePromotionStore struct {
	domainPromotions int
	senderPromotions int
}

func (*fakePromotionStore) Create(context.Context, *frauddomain.ThreatReport) error {
	return nil
}
func (*fakePromotionStore) HighestSeverityForDomain(context.Context, string, time.Time) (string, error) {
	return "critical", nil
}
func (*fakePromotionStore) HighestSeverityForSender(context.Context, string, time.Time) (string, error) {
	return "high", nil
}
func (f *fakePromotionStore) PromoteDomain(
	context.Context,
	string, string, string,
	time.Time,
) error {
	f.domainPromotions++
	return nil
}
func (f *fakePromotionStore) PromoteSender(
	context.Context,
	string, string, string,
	time.Time,
) error {
	f.senderPromotions++
	return nil
}

func TestThreatReportQueueJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 8, 3, 9, 0, 0, 123000000, time.UTC)
	domain := "scam.example"
	report := &frauddomain.ThreatReport{
		ID: "report-1", ThreatType: "url_phishing", Severity: "high",
		URLDomain: &domain, DetectionSource: "ml", SDKVersion: "1.0",
		ReportedAt: now,
	}
	raw, err := encodeThreatReportQueuePayload(report)
	require.NoError(t, err)
	decoded, err := DecodeThreatReportQueuePayload(raw)
	require.NoError(t, err)
	require.Equal(t, report, decoded)
}

func TestThreatReportWorkerPrunesWindowAndPromotesAfterTen(t *testing.T) {
	now := time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)
	domain := "scam.example"
	sender := strings.Repeat("a", 64)
	rdb := newFakeReportRedis()
	store := &fakePromotionStore{}

	old := &frauddomain.ThreatReport{
		ID: "old", ThreatType: "url_phishing", Severity: "critical",
		URLDomain: &domain, ReportedAt: now.Add(-2 * time.Hour),
	}
	require.NoError(t, processThreatReport(
		context.Background(), rdb, store, store, old, now,
	))

	for i := 1; i <= 10; i++ {
		report := &frauddomain.ThreatReport{
			ID:         "recent-" + string(rune('a'+i)),
			ThreatType: "url_phishing", Severity: "high",
			URLDomain: &domain, SenderHash: &sender,
			ReportedAt: now.Add(-time.Duration(i) * time.Minute),
		}
		require.NoError(t, processThreatReport(
			context.Background(), rdb, store, store, report, now,
		))
	}
	require.Zero(t, store.domainPromotions)
	require.Zero(t, store.senderPromotions)

	eleventh := &frauddomain.ThreatReport{
		ID: "recent-11", ThreatType: "url_phishing", Severity: "high",
		URLDomain: &domain, SenderHash: &sender, ReportedAt: now,
	}
	require.NoError(t, processThreatReport(
		context.Background(), rdb, store, store, eleventh, now,
	))
	require.Equal(t, 1, store.domainPromotions)
	require.Equal(t, 1, store.senderPromotions)
	require.Len(t, rdb.sets[indicatorRedisKey("fraud:reports", "domain", domain)], 11)
	require.NotEqual(
		t,
		indicatorRedisKey("fraud:reports", "domain", domain),
		indicatorRedisKey("fraud:reports", "sender", sender),
	)
}
