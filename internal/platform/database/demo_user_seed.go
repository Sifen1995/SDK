package database

import (
	"fmt"
	"log"
	"time"

	intentmodel "skykin-platform/internal/intents/model"
	usermodel "skykin-platform/internal/users/model"

	"gorm.io/gorm"
)

const (
	demoFashionExternalID = "demo-fashion-user"
	demoFashionPhone      = "+251911000001"
	demoFashionCohortSize = 12 // users that qualify for fashion_interest classification
)

// seedDemoFashionUser inserts demo SDK users with phone numbers and sustained
// fashion_interest rows so intent-consistency analysis produces a segment candidate.
func seedDemoFashionUser(db *gorm.DB) {
	seedFashionCohort(db)
}

func seedFashionCohort(db *gorm.DB) {
	now := time.Now().UTC()
	phones := []string{
		"+251911000001", "+251911000002", "+251911000003", "+251911000004",
		"+251911000005", "+251911000006", "+251911000007", "+251911000008",
		"+251911000009", "+251911000010", "+251911000011", "+251911000012",
	}

	for i := 0; i < demoFashionCohortSize; i++ {
		extID := demoFashionExternalID
		if i > 0 {
			extID = fmt.Sprintf("demo-fashion-user-%02d", i+1)
		}
		phone := phones[i]
		user := usermodel.Users{ExternalUserID: extID, PhoneNumber: &phone}
		if err := db.Where("external_user_id = ?", extID).
			Attrs(usermodel.Users{PhoneNumber: &phone}).
			FirstOrCreate(&user).Error; err != nil {
			log.Printf("demo fashion user seed (non-fatal): %v", err)
			continue
		}
		seedFashionIntentsForUser(db, user.ID, now, i)
	}
	log.Printf("demo fashion cohort seeded: %d users (external ids demo-fashion-user … demo-fashion-user-%02d)",
		demoFashionCohortSize, demoFashionCohortSize)
}

func seedFashionIntentsForUser(db *gorm.DB, userID string, now time.Time, offset int) {
	var existing int64
	db.Model(&intentmodel.Intent{}).
		Where("user_id = ? AND intent_name = ?", userID, "fashion_interest").
		Count(&existing)
	if existing >= 6 {
		return
	}

	// 6 distinct days within lookback; last activity 1–3 days ago (within max_age_days=7).
	confidence := 0.82 + float64(offset%5)*0.02 // 0.82 – 0.90
	for day := 1; day <= 6; day++ {
		createdAt := now.AddDate(0, 0, -(day + offset%3)).Truncate(24 * time.Hour).Add(time.Duration(9+day) * time.Hour)
		if err := db.Create(&intentmodel.Intent{
			UserID: userID, IntentName: "fashion_interest",
			Confidence: confidence, CreatedAt: createdAt,
		}).Error; err != nil {
			log.Printf("demo fashion intent seed (non-fatal): %v", err)
		}
	}
}
