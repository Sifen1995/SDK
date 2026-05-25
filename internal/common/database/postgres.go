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

	return db.AutoMigrate(
		&usermodel.Users{},
		&eventmodel.Event{},
		&intentmodel.Intent{},
		&rewardmodel.RewardRule{},
		&rewardmodel.Reward{},
		&authmodel.Developer{},
		&authmodel.Application{},
		&authmodel.APIKey{},
	)
}
