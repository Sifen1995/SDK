package database

import (
	"fmt"
	"log"
	"time"

	"skykin-platform/configs"
	advertisermodel "skykin-platform/internal/ad_portal/model"
	audiencemodel "skykin-platform/internal/audience/model"
	authmodel "skykin-platform/internal/auth/model"
	billingmodel "skykin-platform/internal/billing/model"
	campaignmodel "skykin-platform/internal/campaigns/model"
	deliverymodel "skykin-platform/internal/delivery/model"
	eventpersistence "skykin-platform/internal/events/infrastructure/persistence"
	intentmodel "skykin-platform/internal/intents/model"
	rewardmodel "skykin-platform/internal/rewards/model"
	usermodel "skykin-platform/internal/users/model"

	"gorm.io/datatypes"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	if err := db.AutoMigrate(
		&usermodel.Users{},
		&eventpersistence.EventRecord{},
		&intentmodel.Intent{},
		&rewardmodel.RewardRule{},
		&rewardmodel.Reward{},
		&authmodel.Developer{},
		&authmodel.Application{},
		&authmodel.APIKey{},
		&advertisermodel.Role{},
		&advertisermodel.Advertiser{},
		&advertisermodel.PortalUser{},
		&campaignmodel.Campaign{},
		&campaignmodel.DeliveryLog{},
		&deliverymodel.DeliveryJob{},
		&billingmodel.Channel{},
		&billingmodel.SubscriptionPlan{},
		&billingmodel.AdvertiserSubscription{},
		&billingmodel.BillingRate{},
		&billingmodel.BillingEvent{},
		&billingmodel.Invoice{},
		&audiencemodel.AudienceSegment{},
		&audiencemodel.SegmentPurchase{},
	); err != nil {
		return err
	}

	alignAdPortalSchema(db)
	seedPortalRoles(db)
	seedRewardRules(db)
	seedBillingCatalog(db)
	seedAudienceSegments(db)
	return nil
}

func seedPortalRoles(db *gorm.DB) {
	roles := []advertisermodel.Role{
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
	rules := []rewardmodel.RewardRule{
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
	channels := []billingmodel.Channel{
		{Code: "IN_APP_BANNER", Name: "In-App Banner", Description: "Banner ads shown inside host apps"},
		{Code: "PUSH", Name: "Push Notification", Description: "Push notification delivery"},
		{Code: "SMS_PLUS", Name: "SMS+", Description: "Rich SMS with image and CTA", IsPremium: true},
		{Code: "NATIVE_FEED", Name: "Native Feed", Description: "Native in-feed ad units"},
	}
	for _, ch := range channels {
		db.Where("code = ?", ch.Code).FirstOrCreate(&ch)
	}

	plans := []billingmodel.SubscriptionPlan{
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
	segments := []audiencemodel.AudienceSegment{
		{
			Name: "Fashion Enthusiasts", Description: "Users showing strong fashion and lifestyle purchase intent",
			TopIntentSignals: datatypes.JSON(`["fashion_interest"]`), ApproximateSize: 12500, EstimatedCPM: 4.50,
			AvailableFrom: now, IsActive: true,
		},
		{
			Name: "Crypto & Fintech", Description: "Users interested in crypto trading and fintech products",
			TopIntentSignals: datatypes.JSON(`["crypto_interest","fintech_interest"]`), ApproximateSize: 8200, EstimatedCPM: 6.00,
			AvailableFrom: now, IsActive: true,
		},
		{
			Name: "Food & Dining", Description: "Food delivery and restaurant discovery intent",
			TopIntentSignals: datatypes.JSON(`["food_interest"]`), ApproximateSize: 15000, EstimatedCPM: 3.25,
			AvailableFrom: now, IsActive: true,
		},
		{
			Name: "Mobile Gamers", Description: "Gaming and in-app engagement intent",
			TopIntentSignals: datatypes.JSON(`["gaming_interest"]`), ApproximateSize: 21000, EstimatedCPM: 2.75,
			AvailableFrom: now, IsActive: true,
		},
		{
			Name: "Lifelong Learners", Description: "Education and upskilling intent",
			TopIntentSignals: datatypes.JSON(`["education_interest"]`), ApproximateSize: 9800, EstimatedCPM: 3.80,
			AvailableFrom: now, IsActive: true,
		},
		{
			Name: "Broad Reach", Description: "General engagement across mixed verticals",
			TopIntentSignals: datatypes.JSON(`["general_interest","fashion_interest","food_interest"]`), ApproximateSize: 45000, EstimatedCPM: 1.50,
			AvailableFrom: now, IsActive: true,
		},
	}
	for _, seg := range segments {
		db.Where("name = ?", seg.Name).FirstOrCreate(&seg)
	}
	log.Println("audience segments seeded")
}
