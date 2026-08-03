package persistance

import (
	"context"
	"strings"
	"testing"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newReportTestRepository(t *testing.T) (*gorm.DB, *ReportRepository) {
	t.Helper()
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE threat_reports (
		id TEXT PRIMARY KEY, threat_type TEXT NOT NULL, severity TEXT NOT NULL,
		sender_hash TEXT NULL, url_domain TEXT NULL, detection_source TEXT NOT NULL,
		sdk_version TEXT NOT NULL, reported_at DATETIME NOT NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE blocked_domains (
		domain TEXT PRIMARY KEY, threat_type TEXT NOT NULL, severity TEXT NOT NULL,
		source TEXT NOT NULL, status TEXT NOT NULL, created_at DATETIME NOT NULL,
		expires_at DATETIME NULL, updated_at DATETIME NOT NULL
	)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE blocked_senders (
		sender_hash TEXT PRIMARY KEY, threat_type TEXT NOT NULL, severity TEXT NOT NULL,
		source TEXT NOT NULL, status TEXT NOT NULL, created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	)`).Error)
	return db, NewReportRepository(db)
}

func TestReportRepositoryCreatesAndFindsHighestWindowSeverity(t *testing.T) {
	_, repository := newReportTestRepository(t)
	now := time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)
	domain := "scam.example"
	for id, value := range map[string]struct {
		severity string
		at       time.Time
	}{
		"old-critical": {"critical", now.Add(-2 * time.Hour)},
		"recent-low":   {"low", now.Add(-30 * time.Minute)},
		"recent-high":  {"high", now.Add(-10 * time.Minute)},
	} {
		report := &frauddomain.ThreatReport{
			ID: id, ThreatType: "url_phishing", Severity: value.severity,
			URLDomain: &domain, DetectionSource: "ml", SDKVersion: "1.0",
			ReportedAt: value.at,
		}
		require.NoError(t, repository.Create(context.Background(), report))
	}

	severity, err := repository.HighestSeverityForDomain(
		context.Background(), domain, now.Add(-time.Hour),
	)
	require.NoError(t, err)
	require.Equal(t, "high", severity)
}

func TestReportRepositoryPromotesAndDoesNotDowngrade(t *testing.T) {
	db, repository := newReportTestRepository(t)
	now := time.Date(2026, 8, 3, 9, 0, 0, 0, time.UTC)

	require.NoError(t, repository.PromoteDomain(
		context.Background(), "scam.example", "url_phishing", "critical", now,
	))
	require.NoError(t, repository.PromoteDomain(
		context.Background(), "scam.example", "url_phishing", "low", now.Add(time.Minute),
	))
	var domain BlockedDomainsRow
	require.NoError(t, db.First(&domain, "domain = ?", "scam.example").Error)
	require.Equal(t, "critical", domain.Severity)
	require.Equal(t, "community_report", domain.Source)
	require.Equal(t, frauddomain.StatusActive, domain.Status)
	require.WithinDuration(t, now.AddDate(0, 0, 30).Add(time.Minute), domain.ExpiresAt.Time, time.Second)

	sender := strings.Repeat("a", 64)
	require.NoError(t, repository.PromoteSender(
		context.Background(), sender, "financial_scam", "high", now,
	))
	require.NoError(t, repository.PromoteSender(
		context.Background(), sender, "financial_scam", "medium", now.Add(time.Minute),
	))
	var senderRow BlockedSendersRow
	require.NoError(t, db.First(&senderRow, "sender_hash = ?", sender).Error)
	require.Equal(t, "high", senderRow.Severity)
	require.Equal(t, "community_report", senderRow.Source)
}
