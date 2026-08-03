package persistance

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSyncTestRepository(t *testing.T) (*gorm.DB, *SyncRepository) {
	t.Helper()
	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
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
	require.NoError(t, db.Exec(`CREATE TABLE scam_patterns (
		id TEXT PRIMARY KEY, pattern_type TEXT NOT NULL, pattern_value TEXT NOT NULL,
		threat_category TEXT NOT NULL, confidence REAL NOT NULL, language TEXT NOT NULL,
		is_active BOOLEAN NOT NULL, updated_at DATETIME NOT NULL
	)`).Error)
	return db, NewSyncRepository(db)
}

func TestSyncRepositoryFullSnapshotFiltersInactiveAndExpired(t *testing.T) {
	db, repository := newSyncTestRepository(t)
	now := time.Date(2026, 8, 3, 8, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	require.NoError(t, db.Create(&[]BlockedDomainsRow{
		{Domain: "active.example", ThreatType: "url_phishing", Severity: "high", Source: "test", Status: frauddomain.StatusActive, ExpiresAt: nullableTime(future), CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Minute)},
		{Domain: "expired.example", ThreatType: "url_phishing", Severity: "high", Source: "test", Status: frauddomain.StatusActive, ExpiresAt: nullableTime(expired), CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Minute)},
		{Domain: "revoked.example", ThreatType: "url_phishing", Severity: "high", Source: "test", Status: frauddomain.StatusRevoked, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Minute)},
	}).Error)
	require.NoError(t, db.Create(&[]BlockedSendersRow{
		{SenderHash: "active-hash", ThreatType: "financial_scam", Severity: "high", Source: "test", Status: frauddomain.StatusActive, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Minute)},
		{SenderHash: "revoked-hash", ThreatType: "financial_scam", Severity: "high", Source: "test", Status: frauddomain.StatusRevoked, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Minute)},
	}).Error)
	require.NoError(t, db.Create(&[]ScamPatternsRow{
		{ID: "active-pattern", PatternType: "regex", PatternValue: "urgent", ThreatCategory: "scam", Confidence: .9, Language: "en", IsActive: true, UpdatedAt: now.Add(-time.Minute)},
		{ID: "inactive-pattern", PatternType: "regex", PatternValue: "old", ThreatCategory: "scam", Confidence: .8, Language: "en", IsActive: false, UpdatedAt: now.Add(-time.Minute)},
	}).Error)

	result, err := repository.Sync(context.Background(), nil, now)
	require.NoError(t, err)
	require.False(t, result.IsDelta)
	require.Equal(t, now, result.NextCursor)
	require.Len(t, result.BlockedDomains, 1)
	require.Equal(t, "active.example", result.BlockedDomains[0].Domain)
	require.Len(t, result.BlockedSenders, 1)
	require.Equal(t, "active-hash", result.BlockedSenders[0].SenderHash)
	require.Len(t, result.ScamPatterns, 1)
	require.Equal(t, "active-pattern", result.ScamPatterns[0].ID)
}

func TestSyncRepositoryDeltaIncludesTombstonesAndHonorsBounds(t *testing.T) {
	db, repository := newSyncTestRepository(t)
	since := time.Date(2026, 8, 3, 6, 0, 0, 0, time.UTC)
	until := since.Add(2 * time.Hour)

	require.NoError(t, db.Create(&[]BlockedDomainsRow{
		{Domain: "old.example", ThreatType: "url_phishing", Severity: "high", Source: "test", Status: frauddomain.StatusActive, CreatedAt: since.Add(-time.Hour), UpdatedAt: since},
		{Domain: "revoked.example", ThreatType: "url_phishing", Severity: "critical", Source: "test", Status: frauddomain.StatusRevoked, CreatedAt: since.Add(-time.Hour), UpdatedAt: since.Add(time.Hour)},
		{Domain: "expired-in-window.example", ThreatType: "url_phishing", Severity: "high", Source: "test", Status: frauddomain.StatusActive, CreatedAt: since.Add(-time.Hour), ExpiresAt: nullableTime(since.Add(time.Hour)), UpdatedAt: since},
		{Domain: "future.example", ThreatType: "url_phishing", Severity: "high", Source: "test", Status: frauddomain.StatusActive, CreatedAt: since, UpdatedAt: until.Add(time.Minute)},
	}).Error)
	require.NoError(t, db.Create(&BlockedSendersRow{
		SenderHash: "revoked-hash", ThreatType: "financial_scam", Severity: "high", Source: "test", Status: frauddomain.StatusRevoked, CreatedAt: since, UpdatedAt: since.Add(time.Hour),
	}).Error)
	require.NoError(t, db.Create(&ScamPatternsRow{
		ID: "inactive-pattern", PatternType: "regex", PatternValue: "old", ThreatCategory: "scam", Confidence: .8, Language: "en", IsActive: false, UpdatedAt: since.Add(time.Hour),
	}).Error)

	result, err := repository.Sync(context.Background(), &since, until)
	require.NoError(t, err)
	require.True(t, result.IsDelta)
	require.Len(t, result.BlockedDomains, 2)
	domains := make(map[string]frauddomain.BlockedDomain, len(result.BlockedDomains))
	for _, domain := range result.BlockedDomains {
		domains[domain.Domain] = domain
	}
	require.Equal(t, frauddomain.StatusRevoked, domains["revoked.example"].Status)
	require.Equal(t, frauddomain.StatusRevoked, domains["expired-in-window.example"].Status)
	require.Len(t, result.BlockedSenders, 1)
	require.Equal(t, frauddomain.StatusRevoked, result.BlockedSenders[0].Status)
	require.Len(t, result.ScamPatterns, 1)
	require.False(t, result.ScamPatterns[0].IsActive)
}

func nullableTime(value time.Time) sql.NullTime {
	return sql.NullTime{Time: value, Valid: true}
}
