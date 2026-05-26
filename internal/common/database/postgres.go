package database

import (
	"fmt"
	"log"
	"skykin-platform/configs"
	authmodel "skykin-platform/internal/auth/model"
	eventmodel "skykin-platform/internal/events/model"
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
		&eventmodel.Event{},
		&intentmodel.Intent{},
		&rewardmodel.RewardRule{},
		&rewardmodel.Reward{},
		&authmodel.Developer{},
		&authmodel.Application{},
		&authmodel.APIKey{},
	); err != nil {
		return err
	}

	seedRewardRules(db)
	return nil
}

func seedRewardRules(db *gorm.DB) {
	rules := []rewardmodel.RewardRule{
		{IntentName: "coffee_interest", RewardType: "cashback", Amount: 20.00, Currency: "ETB", Message: "You earned 20 ETB cashback for your coffee passion!", IsActive: true},
		{IntentName: "crypto_interest", RewardType: "coins", Amount: 50.00, Currency: "FLIP_COINS", Message: "Crypto enthusiast! You earned 50 Flip Coins!", IsActive: true},
		{IntentName: "fashion_interest", RewardType: "cashback", Amount: 15.00, Currency: "ETB", Message: "Stylish! Here is 15 ETB store credit for your next look.", IsActive: true},
		{IntentName: "abandoned_cart", RewardType: "discount", Amount: 10.00, Currency: "PERCENT", Message: "We noticed you left something behind! Here is a 10% discount.", IsActive: true},
		{IntentName: "signup_intent", RewardType: "points", Amount: 100.00, Currency: "POINTS", Message: "Welcome to Skykin! You earned 100 loyalty points.", IsActive: true},
	}

	for _, rule := range rules {
		db.Where("intent_name = ?", rule.IntentName).FirstOrCreate(&rule)
	}
	log.Println("reward rules seeded")
}
