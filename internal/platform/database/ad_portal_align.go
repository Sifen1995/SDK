package database

import (
	"log"

	"gorm.io/gorm"
)

// alignAdPortalSchema removes legacy columns from advertisers when an older AutoMigrate
// left email/password on advertisers instead of portal_users.
func alignAdPortalSchema(db *gorm.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			slug VARCHAR(50) NOT NULL UNIQUE,
			display_name VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO roles (slug, display_name) VALUES
			('operator_admin', 'Operator Admin'),
			('advertiser', 'Advertiser'),
			('read_only_analyst', 'Read-Only Analyst')
		ON CONFLICT (slug) DO NOTHING`,
		`ALTER TABLE advertisers DROP COLUMN IF EXISTS email`,
		`ALTER TABLE advertisers DROP COLUMN IF EXISTS password_hash`,
		`ALTER TABLE advertisers DROP COLUMN IF EXISTS api_key`,
		`ALTER TABLE advertisers DROP COLUMN IF EXISTS role`,
		`ALTER TABLE advertisers DROP COLUMN IF EXISTS contact_name`,
		`ALTER TABLE advertisers DROP COLUMN IF EXISTS is_active`,
		`ALTER TABLE advertisers ADD COLUMN IF NOT EXISTS company_name VARCHAR(255)`,
		`UPDATE advertisers SET company_name = 'Migrated Company' WHERE company_name IS NULL`,
	}

	var advertisersExists bool
	db.Raw(`SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'advertisers'
	)`).Scan(&advertisersExists)
	if !advertisersExists {
		return
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("ad portal schema align (non-fatal): %v", err)
		}
	}

	// company_name NOT NULL (ignore if already set)
	_ = db.Exec(`ALTER TABLE advertisers ALTER COLUMN company_name SET NOT NULL`).Error
	_ = db.Exec(`ALTER TABLE campaigns DROP COLUMN IF EXISTS application_id`).Error
	_ = db.Exec(`ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS destination_url TEXT`).Error
	_ = db.Exec(`ALTER TABLE campaigns DROP COLUMN IF EXISTS billing_model`).Error
	_ = db.Exec(`ALTER TABLE campaigns DROP COLUMN IF EXISTS creative_format`).Error
	_ = db.Exec(`ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(20) NOT NULL DEFAULT 'pending'`).Error
	_ = db.Exec(`ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderation_notes TEXT`).Error
	_ = db.Exec(`ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderated_at TIMESTAMPTZ`).Error
	_ = db.Exec(`ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS moderated_by UUID`).Error
	_ = db.Exec(`UPDATE campaigns SET moderation_status = 'approved' WHERE is_active = true AND moderation_status = 'pending'`).Error
	log.Println("ad portal schema aligned (advertisers = company only)")
}
