package database

import (
	_ "embed"
	"fmt"
	"log"
	"time"

	"skykin-platform/configs"
	adportalpersistence "skykin-platform/internal/ad_portal/infrastructure/persistence"
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

func ConnectDB(cfg *configs.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm connection: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging Postgres: %v", err)
	}

	log.Println("database connected")
	return db, nil
}

func Migrate(db *gorm.DB) error {
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	alignPersistenceTimestamps(db)

	if err := db.AutoMigrate(
		&userpersistence.UserRow{},
		&eventpersistence.EventRecord{},
		&intentpersistence.IntentRow{},
		&rewardpersistence.RewardRuleRow{},
		&rewardpersistence.RewardRow{},
		&authpersistence.DeveloperRow{},
		&authpersistence.ApplicationRow{},
		&authpersistence.APIKeyRow{},
		&adportalpersistence.RoleRow{},
		&adportalpersistence.AdvertiserRow{},
		&adportalpersistence.PortalUserRow{},
		&campaignpersistence.CampaignRow{},
		&campaignpersistence.DeliveryLogRow{},
		&deliverypersistence.DeliveryJobRow{},
		&billingpersistence.ChannelRow{},
		&billingpersistence.SubscriptionPlanRow{},
		&billingpersistence.AdvertiserSubscriptionRow{},
		&billingpersistence.BillingRateRow{},
		&billingpersistence.BillingEventRow{},
		&billingpersistence.InvoiceRow{},
		&audiencepersistence.AudienceSegmentRow{},
		&audiencepersistence.SegmentPurchaseRow{},
	); err != nil {
		return err
	}

	alignAdPortalSchema(db)
	alignSegmentClassificationSchema(db)
	if err := applyPermissionsMigration(db); err != nil {
		return fmt.Errorf("permissions migration: %w", err)
	}
	seedPortalRoles(db)
	seedRewardRules(db)
	seedBillingCatalog(db)
	seedAudienceSegments(db)
	seedDemoFashionUser(db)
	return nil
}

// alignPersistenceTimestamps backfills null audit columns before AutoMigrate adds NOT NULL constraints.
func alignPersistenceTimestamps(db *gorm.DB) {
	stmts := []string{
		`UPDATE reward_rules SET created_at = NOW() WHERE created_at IS NULL`,
		`UPDATE rewards SET created_at = NOW() WHERE created_at IS NULL`,
	}
	for _, stmt := range stmts {
		_ = db.Exec(stmt).Error
	}
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
