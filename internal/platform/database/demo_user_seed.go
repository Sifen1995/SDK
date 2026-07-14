package database

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	consentpersistence "skykin-platform/internal/consent/infrastructure/persistence"
	intentpersistence "skykin-platform/internal/intents/infrastructure/persistence"
	userpersistence "skykin-platform/internal/users/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const demoFashionCohortSize = 12

// seedDemoFashionUser inserts demo SDK users (random bigint ids) with mappings
// and sustained fashion_interest rows so intent-consistency analysis works.
func seedDemoFashionUser(db *gorm.DB) {
	seedFashionCohort(db)
}

func seedFashionCohort(db *gorm.DB) {
	var existing int64
	if err := db.Model(&userpersistence.UserRow{}).Count(&existing).Error; err == nil && existing > 0 {
		return
	}

	now := time.Now().UTC()
	for i := 0; i < demoFashionCohortSize; i++ {
		id, err := randomDemoUserID()
		if err != nil {
			log.Printf("demo fashion user seed (non-fatal): %v", err)
			continue
		}
		user := userpersistence.UserRow{ID: id}
		if err := db.Create(&user).Error; err != nil {
			log.Printf("demo fashion user seed (non-fatal): %v", err)
			continue
		}
		mapping := consentpersistence.PseudonymousMappingRow{
			UserID:         user.ID,
			PseudonymousID: uuid.New(),
		}
		if err := db.Create(&mapping).Error; err != nil {
			log.Printf("demo fashion mapping seed (non-fatal): %v", err)
		}
		consent := consentpersistence.ConsentRow{
			UserID:       user.ID,
			ConsentLevel: "individual",
			IsActive:     true,
			GrantedAt:    &now,
			SDKVersion:   "demo",
		}
		if err := db.Create(&consent).Error; err != nil {
			log.Printf("demo fashion consent seed (non-fatal): %v", err)
		}
		seedFashionIntentsForUser(db, fmt.Sprintf("%d", user.ID), now, i)
	}
	log.Printf("demo fashion cohort seeded: %d users (bigint ids + pseudonymous mappings)", demoFashionCohortSize)
}

func seedFashionIntentsForUser(db *gorm.DB, userID string, now time.Time, offset int) {
	var existing int64
	db.Model(&intentpersistence.IntentRow{}).
		Where("user_id = ? AND intent_name = ?", userID, "fashion_interest").
		Count(&existing)
	if existing >= 6 {
		return
	}

	confidence := 0.82 + float64(offset%5)*0.02
	for day := 1; day <= 6; day++ {
		createdAt := now.AddDate(0, 0, -(day + offset%3)).Truncate(24 * time.Hour).Add(time.Duration(9+day) * time.Hour)
		if err := db.Create(&intentpersistence.IntentRow{
			UserID: userID, IntentName: "fashion_interest",
			Confidence: confidence, CreatedAt: createdAt,
		}).Error; err != nil {
			log.Printf("demo fashion intent seed (non-fatal): %v", err)
		}
	}
}

func randomDemoUserID() (int64, error) {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0, err
	}
	n := int64(binary.BigEndian.Uint64(buf[:]) & 0x7fffffffffffffff)
	if n == 0 {
		n = 1
	}
	return n, nil
}
