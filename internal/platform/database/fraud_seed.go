package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"time"

	frauddomain "skykin-platform/internal/fraud/domain"
	fraudpersistance "skykin-platform/internal/fraud/infrastructure/persistance"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// seedFraudIntelligence inserts demo blocklist / pattern rows so GET /api/v1/sync
// returns usable data for Flutter SQLite cache testing. Idempotent via primary keys.
func seedFraudIntelligence(db *gorm.DB) {
	now := time.Now().UTC()
	expires := now.AddDate(0, 3, 0)

	domains := []fraudpersistance.BlockedDomainsRow{
		{
			Domain: "telebirr-verify-kyc.xyz", ThreatType: "brand_impersonation",
			Severity: "critical", Source: "manual_review", Status: frauddomain.StatusActive,
			ExpiresAt: sql.NullTime{Time: expires, Valid: true}, CreatedAt: now, UpdatedAt: now,
		},
		{
			Domain: "cbe-birr-unblock.top", ThreatType: "financial_scam",
			Severity: "critical", Source: "auto_detected", Status: frauddomain.StatusActive,
			ExpiresAt: sql.NullTime{Time: expires, Valid: true}, CreatedAt: now, UpdatedAt: now,
		},
		{
			Domain: "ethiotel-bonus-5g.sbs", ThreatType: "url_phishing",
			Severity: "high", Source: "community_report", Status: frauddomain.StatusActive,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			Domain: "bit.ly/ethio-bonus-claim", ThreatType: "url_phishing",
			Severity: "high", Source: "manual_review", Status: frauddomain.StatusActive,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			Domain: "192.168.1.105", ThreatType: "url_phishing",
			Severity: "critical", Source: "auto_detected", Status: frauddomain.StatusActive,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			Domain: "old-telebirr-scam.cfd", ThreatType: "brand_impersonation",
			Severity: "medium", Source: "manual_review", Status: frauddomain.StatusRevoked,
			CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour),
		},
	}

	senders := []fraudpersistance.BlockedSendersRow{
		{
			SenderHash: hashSender("+251911000001"), ThreatType: "financial_scam",
			Severity: "critical", Source: "manual_review", Status: frauddomain.StatusActive,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			SenderHash: hashSender("+251922000002"), ThreatType: "brand_impersonation",
			Severity: "high", Source: "community_report", Status: frauddomain.StatusActive,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			SenderHash: hashSender("994"), ThreatType: "url_phishing",
			Severity: "high", Source: "auto_detected", Status: frauddomain.StatusActive,
			CreatedAt: now, UpdatedAt: now,
		},
		{
			SenderHash: hashSender("+251933000003"), ThreatType: "financial_scam",
			Severity: "medium", Source: "manual_review", Status: frauddomain.StatusRevoked,
			CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-12 * time.Hour),
		},
	}

	patterns := []fraudpersistance.ScamPatternsRow{
		{
			ID: "urgent-telebirr-verify-en", PatternType: "regex",
			PatternValue:   `(?i)(urgent|suspended|locked).{0,40}telebirr.{0,40}(verify|update|click)`,
			ThreatCategory: "brand_impersonation", Confidence: 0.95, Language: "en",
			IsActive: true, UpdatedAt: now,
		},
		{
			ID: "telebirr-blocked-am", PatternType: "keyword_combo",
			PatternValue:   "ማስጠንቀቂያ|ታግዷል|Telebirr|ያረጋግጡ",
			ThreatCategory: "brand_impersonation", Confidence: 0.92, Language: "am",
			IsActive: true, UpdatedAt: now,
		},
		{
			ID: "cbe-account-locked-en", PatternType: "regex",
			PatternValue:   `(?i)cbe.{0,30}(locked|blocked|suspicious).{0,40}(verify|restore|click)`,
			ThreatCategory: "financial_scam", Confidence: 0.94, Language: "en",
			IsActive: true, UpdatedAt: now,
		},
		{
			ID: "ethio-bonus-claim-url", PatternType: "url_pattern",
			PatternValue:   `(?i)(bonus|reward|claim).{0,20}(http|www\.)`,
			ThreatCategory: "url_phishing", Confidence: 0.88, Language: "any",
			IsActive: true, UpdatedAt: now,
		},
		{
			ID: "ip-literal-http", PatternType: "url_pattern",
			PatternValue:   `https?://\d{1,3}(?:\.\d{1,3}){3}`,
			ThreatCategory: "url_phishing", Confidence: 0.90, Language: "any",
			IsActive: true, UpdatedAt: now,
		},
		{
			ID: "legacy-sms-lottery", PatternType: "keyword_combo",
			PatternValue:   "congratulations|won|lottery|claim prize",
			ThreatCategory: "financial_scam", Confidence: 0.70, Language: "en",
			IsActive: false, UpdatedAt: now.Add(-24 * time.Hour),
		},
	}

	for i := range domains {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&domains[i]).Error; err != nil {
			log.Printf("fraud domain seed (non-fatal): %v", err)
		}
	}
	for i := range senders {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&senders[i]).Error; err != nil {
			log.Printf("fraud sender seed (non-fatal): %v", err)
		}
	}
	for i := range patterns {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&patterns[i]).Error; err != nil {
			log.Printf("fraud pattern seed (non-fatal): %v", err)
		}
	}

	var domainCount, senderCount, patternCount int64
	_ = db.Model(&fraudpersistance.BlockedDomainsRow{}).Count(&domainCount).Error
	_ = db.Model(&fraudpersistance.BlockedSendersRow{}).Count(&senderCount).Error
	_ = db.Model(&fraudpersistance.ScamPatternsRow{}).Count(&patternCount).Error
	log.Printf(
		"fraud intelligence seeded: %d domains, %d senders, %d patterns",
		domainCount, senderCount, patternCount,
	)
}

func hashSender(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
