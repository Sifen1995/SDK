package database

import (
	"fmt"
	"log"
	"skykin-platform/configs"
	advertisermodel "skykin-platform/internal/advertisers/model"
	campaignmodel "skykin-platform/internal/campaigns/model"
	authmodel "skykin-platform/internal/auth/model"
	eventpersistence "skykin-platform/internal/events/infrastructure/persistence"
	intentmodel "skykin-platform/internal/intents/model"
	rewardmodel "skykin-platform/internal/rewards/model"
	usermodel "skykin-platform/internal/users/model"

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
	); err != nil {
		return err
	}

	alignAdPortalSchema(db)
	seedPortalRoles(db)
	seedRewardRules(db)
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

