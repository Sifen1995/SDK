package database

import (
	_ "embed"
	"fmt"
	"log"
	"time"

	"skykin-platform/configs"
	adportalpersistence "skykin-platform/internal/ad_portal/infrastructure/persistence"
	analyticspersistence "skykin-platform/internal/analytics/infrastructure/persistence"
	audiencepersistence "skykin-platform/internal/audience/infrastructure/persistence"
	authpersistence "skykin-platform/internal/auth/infrastructure/persistence"
	billingpersistence "skykin-platform/internal/billing/infrastructure/persistence"
	campaignpersistence "skykin-platform/internal/campaigns/infrastructure/persistence"
	deliverypersistence "skykin-platform/internal/delivery/infrastructure/persistence"
	eventpersistence "skykin-platform/internal/events/infrastructure/persistence"
	intentpersistence "skykin-platform/internal/intents/infrastructure/persistence"
	rewardpersistence "skykin-platform/internal/rewards/infrastructure/persistence"
	userpersistence "skykin-platform/internal/users/infrastructure/persistence"

	"gorm.io/datatypes"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/20260703120000_permissions.sql
var permissionsMigrationSQL string

//go:embed migrations/20260712153000_users_consent_identity.sql
var usersConsentIdentitySQL string

//go:embed migrations/20260729120000_pseudonymous_identity.sql
var pseudonymousIdentitySQL string

//go:embed migrations/20260730130000_sms_demo_delivery.sql
var smsDemoDeliverySQL string

//go:embed migrations/20260730140000_consents_sms_consented.sql
var consentsSMSConsentedSQL string

//go:embed migrations/20260730150000_demo_sms_recipient_pseudonymous.sql
var demoSMSRecipientPseudonymousSQL string

//go:embed migrations/fraud.sql
var fraudSQL string

//go:embed migrations/geofencing.sql
var geofencingSQL string

//go:embed migrations/20260808120000_analysts.sql
var analystsSQL string

func ConnectDB(cfg *configs.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	const (
		maxAttempts = 30
		retryDelay  = 2 * time.Second
	)

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
		if err != nil {
			lastErr = fmt.Errorf("failed to open gorm connection: %v", err)
		} else {
			sqlDB, err := db.DB()
			if err != nil {
				lastErr = fmt.Errorf("failed to get underlying sql.DB: %v", err)
			} else {
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetMaxOpenConns(100)
				if err := sqlDB.Ping(); err != nil {
					lastErr = fmt.Errorf("error pinging Postgres: %v", err)
				} else {
					log.Println("database connected")
					return db, nil
				}
			}
		}
		if attempt < maxAttempts {
			log.Printf("database not ready (attempt %d/%d): %v", attempt, maxAttempts, lastErr)
			time.Sleep(retryDelay)
		}
	}
	return nil, fmt.Errorf("failed to connect to postgres after %d attempts: %v", maxAttempts, lastErr)
}

func Migrate(db *gorm.DB) error {
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	alignPersistenceTimestamps(db)

	if err := applyUsersConsentIdentityMigration(db); err != nil {
		return fmt.Errorf("users/consent identity migration: %w", err)
	}

	// Must run before AutoMigrate so GORM sees pseudonymous_id and does not re-add user_id.
	if err := db.Exec(pseudonymousIdentitySQL).Error; err != nil {
		return fmt.Errorf("pseudonymous identity migration: %w", err)
	}

	if err := db.Exec(consentsSMSConsentedSQL).Error; err != nil {
		return fmt.Errorf("consents sms_consented migration: %w", err)
	}

	if err := db.Exec(fraudSQL).Error; err != nil {
		return fmt.Errorf("fraud sync migration: %w", err)
	}

	if err := db.AutoMigrate(
		&userpersistence.UserRow{},
		&eventpersistence.EventRecord{},
		&intentpersistence.IntentRow{},
		&analyticspersistence.IntentAggregateCountRow{},
		&rewardpersistence.RewardRuleRow{},
		&rewardpersistence.RewardRow{},
		&authpersistence.DeveloperRow{},
		&authpersistence.ApplicationRow{},
		&authpersistence.APIKeyRow{},
		&adportalpersistence.RoleRow{},
		&adportalpersistence.AdvertiserRow{},
		&adportalpersistence.AnalystRow{},
		&adportalpersistence.PortalUserRow{},
		&campaignpersistence.CampaignRow{},
		&deliverypersistence.DeliveryJobRow{},
		&deliverypersistence.DeliveryLogRow{},
		&billingpersistence.ChannelRow{},
		&billingpersistence.SubscriptionPlanRow{},
		&billingpersistence.AdvertiserSubscriptionRow{},
		&billingpersistence.BillingRateRow{},
		&billingpersistence.BillingEventRow{},
		&billingpersistence.InvoiceRow{},
		&audiencepersistence.AudienceSegmentRow{},
		&audiencepersistence.SegmentPurchaseRow{},
		// consents + pseudonymous_mappings are created by applyUsersConsentIdentityMigration
		// (explicit unique constraint names). Do not AutoMigrate them — GORM would try to
		// DROP uni_pseudonymous_mappings_* constraints that do not exist.
		// demo_sms_recipients + sms_send_attempts are also created by explicit SQL migration
		// so GORM does not attempt to rewrite their unique constraints on startup.
	); err != nil {
		return err
	}

	// Runs after AutoMigrate: sms_send_attempts has a foreign key to campaigns(id) and
	// demo_sms_recipients to users(id), both of which AutoMigrate creates. On a fresh
	// database these statements fail with 42P01 if they execute earlier.
	if err := db.Exec(smsDemoDeliverySQL).Error; err != nil {
		return fmt.Errorf("sms demo delivery migration: %w", err)
	}

	if err := db.Exec(demoSMSRecipientPseudonymousSQL).Error; err != nil {
		return fmt.Errorf("demo sms recipient pseudonymous migration: %w", err)
	}

	// After campaigns + demo_sms_recipients exist (FKs / ALTER COLUMN).
	if err := db.Exec(geofencingSQL).Error; err != nil {
		return fmt.Errorf("geofencing migration: %w", err)
	}
	ensureGeofenceAdvertiserPermission(db)

	if err := db.Exec(analystsSQL).Error; err != nil {
		return fmt.Errorf("analysts migration: %w", err)
	}

	alignAdPortalSchema(db)
	if err := alignSegmentClassificationSchema(db); err != nil {
		return err
	}
	ensureIntentAggregateCountsTable(db)
	if err := applyPermissionsMigration(db); err != nil {
		return fmt.Errorf("permissions migration: %w", err)
	}
	seedPortalRoles(db)
	seedRewardRules(db)
	seedBillingCatalog(db)
	seedAudienceSegments(db)
	seedDemoFashionUser(db)
	seedDemoSMSRecipients(db)
	seedFraudIntelligence(db)
	return nil
}

// alignPersistenceTimestamps backfills null audit columns before AutoMigrate adds NOT NULL constraints.
// Skips tables that do not exist yet (fresh volume / first boot).
func alignPersistenceTimestamps(db *gorm.DB) {
	for _, table := range []string{"reward_rules", "rewards"} {
		if !tableExists(db, table) {
			continue
		}
		_ = db.Exec(
			fmt.Sprintf(`UPDATE %s SET created_at = NOW() WHERE created_at IS NULL`, table),
		).Error
	}
}

func tableExists(db *gorm.DB, name string) bool {
	var count int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = ?
	`, name).Scan(&count).Error
	return count > 0
}

func seedPortalRoles(db *gorm.DB) {
	roles := []adportalpersistence.RoleRow{
		{Slug: "operator_admin", DisplayName: "Operator Admin"},
		{Slug: "advertiser", DisplayName: "Advertiser"},
		{Slug: "read_only_analyst", DisplayName: "Read-Only Analyst"},
	}
	for _, role := range roles {
		db.Where("slug = ?", role.Slug).FirstOrCreate(&role)
	}
	log.Println("portal roles seeded")
}

func seedRewardRules(db *gorm.DB) {
	rules := []rewardpersistence.RewardRuleRow{
		{IntentName: "fashion_interest", RewardType: "coins", Amount: 20.00, Currency: "COINS", Message: "Fashion explorer! You earned 20 Coins!", IsActive: true},
		{IntentName: "crypto_interest", RewardType: "coins", Amount: 50.00, Currency: "COINS", Message: "Crypto enthusiast! You earned 50 Coins!", IsActive: true},
		{IntentName: "food_interest", RewardType: "cashback", Amount: 15.00, Currency: "ETB", Message: "Foodie reward: 15 ETB cashback!", IsActive: true},
		{IntentName: "education_interest", RewardType: "points", Amount: 100.00, Currency: "POINTS", Message: "Learner bonus: 100 points!", IsActive: true},
		{IntentName: "gaming_interest", RewardType: "coins", Amount: 30.00, Currency: "COINS", Message: "Gamer reward: 30 Coins!", IsActive: true},
		{IntentName: "fintech_interest", RewardType: "cashback", Amount: 25.00, Currency: "ETB", Message: "Fintech power user: 25 ETB cashback!", IsActive: true},
		{IntentName: "general_interest", RewardType: "points", Amount: 10.00, Currency: "POINTS", Message: "Thanks for engaging! 10 points earned.", IsActive: true},
	}

	for _, rule := range rules {
		db.Where("intent_name = ?", rule.IntentName).FirstOrCreate(&rule)
	}
	log.Println("reward rules seeded")
}

// seedBillingCatalog ensures channels and subscription plans exist (idempotent).
func seedBillingCatalog(db *gorm.DB) {
	channels := []billingpersistence.ChannelRow{
		{Code: "IN_APP_BANNER", Name: "In-App Banner", Description: "Banner ads shown inside host apps"},
		{Code: "PUSH", Name: "Push Notification", Description: "Push notification delivery"},
		{Code: "SMS_PLUS", Name: "SMS+", Description: "Rich SMS with image and CTA", IsPremium: true},
		{Code: "NATIVE_FEED", Name: "Native Feed", Description: "Native in-feed ad units"},
	}
	for _, ch := range channels {
		db.Where("code = ?", ch.Code).FirstOrCreate(&ch)
	}

	plans := []billingpersistence.SubscriptionPlanRow{
		{Name: "Starter", MonthlyFeeETB: 5000, MaxActiveCampaigns: 3, MaxDailyBudgetETB: 500, IncludedImpressions: 10000},
		{Name: "Growth", MonthlyFeeETB: 15000, MaxActiveCampaigns: 10, MaxDailyBudgetETB: 2000, IncludedImpressions: 50000, SMSPlusEnabled: true, AudiencemartEnabled: true, CPCDiscountPct: 5},
		{Name: "Enterprise", MonthlyFeeETB: 50000, MaxActiveCampaigns: 100, MaxDailyBudgetETB: 10000, IncludedImpressions: 200000, SMSPlusEnabled: true, AudiencemartEnabled: true, CPCDiscountPct: 15},
	}
	for _, p := range plans {
		db.Where("name = ?", p.Name).FirstOrCreate(&p)
	}
	log.Println("billing catalog seeded")
}

// seedAudienceSegments inserts catalog Audiencemart cohorts (idempotent by name).
func seedAudienceSegments(db *gorm.DB) {
	now := time.Now().UTC()
	segments := []struct {
		name        string
		description string
		signals     string
		size        int
		cpm         float64
	}{
		{"Fashion Enthusiasts", "Users showing strong fashion and lifestyle purchase intent", `["fashion_interest"]`, 12500, 4.50},
		{"Crypto & Fintech", "Users interested in crypto trading and fintech products", `["crypto_interest","fintech_interest"]`, 8200, 6.00},
		{"Food & Dining", "Food delivery and restaurant discovery intent", `["food_interest"]`, 15000, 3.25},
		{"Mobile Gamers", "Gaming and in-app engagement intent", `["gaming_interest"]`, 21000, 2.75},
		{"Lifelong Learners", "Education and upskilling intent", `["education_interest"]`, 9800, 3.80},
		{"Broad Reach", "General engagement across mixed verticals", `["general_interest","fashion_interest","food_interest"]`, 45000, 1.50},
	}
	for _, seg := range segments {
		row := audiencepersistence.AudienceSegmentRow{
			Name:             seg.name,
			Description:      seg.description,
			TopIntentSignals: datatypes.JSON(seg.signals),
			ApproximateSize:  seg.size,
			EstimatedCPM:     seg.cpm,
			AvailableFrom:    now,
			IsActive:         true,
		}
		db.Where("name = ?", seg.name).FirstOrCreate(&row)
	}
	log.Println("audience segments seeded")
}

// ensureIntentAggregateCountsTable guarantees the anonymous aggregate rollup table + unique constraint.
func ensureIntentAggregateCountsTable(db *gorm.DB) {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS intent_aggregate_counts (
			id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
			intent_name    VARCHAR(100)  NOT NULL,
			date_bucket    DATE          NOT NULL DEFAULT CURRENT_DATE,
			signal_count   INTEGER       NOT NULL DEFAULT 0,
			weighted_count NUMERIC(10,2) NOT NULL DEFAULT 0.00,
			CONSTRAINT uq_intent_date UNIQUE (intent_name, date_bucket)
		)
	`).Error; err != nil {
		log.Printf("ensure intent_aggregate_counts (non-fatal): %v", err)
		return
	}
	_ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_agg_intent_date ON intent_aggregate_counts (intent_name, date_bucket DESC)`).Error
}

func applyPermissionsMigration(db *gorm.DB) error {
	var count int64
	if err := db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'rbac_permissions'
	`).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if err := db.Exec(permissionsMigrationSQL).Error; err != nil {
		return err
	}
	log.Println("permissions schema migrated and seeded")
	return nil
}

// applyUsersConsentIdentityMigration rebuilds users without phone/external_user_id
// (bigint random ids) and consent/mapping tables. Runs once when legacy schema is detected.
func applyUsersConsentIdentityMigration(db *gorm.DB) error {
	var externalCols int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'external_user_id'
	`).Scan(&externalCols).Error

	var idType string
	_ = db.Raw(`
		SELECT data_type FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'id'
	`).Scan(&idType).Error

	var usersExists int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'users'
	`).Scan(&usersExists).Error

	alreadyMigrated := usersExists > 0 && externalCols == 0 && idType == "bigint"
	if alreadyMigrated {
		return ensureConsentTables(db)
	}

	// Clear intent rows that referenced legacy uuid user ids before rebuild.
	// On a brand-new database the intents table does not exist yet — skip quietly.
	if tableExists(db, "intents") {
		_ = db.Exec(`TRUNCATE TABLE intents RESTART IDENTITY CASCADE`).Error
	}

	if err := db.Exec(usersConsentIdentitySQL).Error; err != nil {
		return err
	}
	log.Println("users/consent identity schema rebuilt (legacy columns and data removed)")
	return nil
}

// ensureConsentTables creates consents /pseudonymous_mappings if missing without
// touching an already-migrated users table.
func ensureConsentTables(db *gorm.DB) error {
	var mappings int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'pseudonymous_mappings'
	`).Scan(&mappings).Error
	if mappings > 0 {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS consents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			consent_level VARCHAR(20) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			granted_at TIMESTAMPTZ,
			revoked_at TIMESTAMPTZ,
			sdk_version VARCHAR(20) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS pseudonymous_mappings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			pseudonymous_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT uq_pseudonymous_mappings_user UNIQUE (user_id),
			CONSTRAINT uq_pseudonymous_mappings_pseudo UNIQUE (pseudonymous_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_consents_user_id ON consents (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pseudonymous_mappings_user ON pseudonymous_mappings (user_id)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	log.Println("consent tables ensured")
	return nil
}

// ensureGeofenceAdvertiserPermission assigns geofences:manage to advertiser when RBAC already exists.
func ensureGeofenceAdvertiserPermission(db *gorm.DB) {
	if !tableExists(db, "rbac_permissions") || !tableExists(db, "rbac_role_permissions") {
		return
	}
	_ = db.Exec(`
		INSERT INTO rbac_permissions (name, resource, action, description)
		SELECT 'geofences:manage', 'geofences', 'manage', 'Create and manage geofence zones'
		WHERE NOT EXISTS (SELECT 1 FROM rbac_permissions WHERE name = 'geofences:manage')
	`).Error
	_ = db.Exec(`
		INSERT INTO rbac_role_permissions (role_id, permission_id)
		SELECT r.id, p.id
		FROM rbac_roles r
		CROSS JOIN rbac_permissions p
		WHERE r.name = 'advertiser' AND p.name = 'geofences:manage'
		  AND NOT EXISTS (
			SELECT 1 FROM rbac_role_permissions rp
			WHERE rp.role_id = r.id AND rp.permission_id = p.id
		  )
	`).Error
}
