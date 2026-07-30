package database

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	consentpersistence "skykin-platform/internal/consent/infrastructure/persistence"
	deliverypersistence "skykin-platform/internal/delivery/infrastructure/persistence"
	intentpersistence "skykin-platform/internal/intents/infrastructure/persistence"
	userpersistence "skykin-platform/internal/users/infrastructure/persistence"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const demoFashionCohortSize = 12
const demoSDKVersion = "demo"

// seedDemoFashionUser inserts demo SDK users with mappings, consents, and
// fashion intents. Runs ensure on every migrate so mappings always exist.
func seedDemoFashionUser(db *gorm.DB) {
	ensureDemoFashionCohort(db)
}

// ensureDemoFashionCohort guarantees demoFashionCohortSize users tagged with
// sdk_version=demo each have a pseudonymous_mappings row and sms_consented=true.
func ensureDemoFashionCohort(db *gorm.DB) {
	var demoUserIDs []int64
	if err := db.Raw(
		`SELECT DISTINCT user_id FROM consents WHERE sdk_version = ? ORDER BY user_id`,
		demoSDKVersion,
	).Scan(&demoUserIDs).Error; err != nil {
		log.Printf("demo fashion cohort ensure (non-fatal): %v", err)
		return
	}

	now := time.Now().UTC()
	for i, userID := range demoUserIDs {
		ensureDemoUserArtifacts(db, userID, now, i)
	}

	for i := len(demoUserIDs); i < demoFashionCohortSize; i++ {
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
		ensureDemoUserArtifacts(db, user.ID, now, i)
	}

	var mappingCount int64
	_ = db.Raw(`
		SELECT COUNT(*) FROM pseudonymous_mappings m
		INNER JOIN consents c ON c.user_id = m.user_id AND c.sdk_version = ?
	`, demoSDKVersion).Scan(&mappingCount).Error
	log.Printf("demo fashion cohort ready: %d demo consents with mappings", mappingCount)
}

func ensureDemoUserArtifacts(db *gorm.DB, userID int64, now time.Time, offset int) {
	var mapping consentpersistence.PseudonymousMappingRow
	err := db.Where("user_id = ?", userID).First(&mapping).Error
	if err == gorm.ErrRecordNotFound {
		mapping = consentpersistence.PseudonymousMappingRow{
			UserID:         userID,
			PseudonymousID: uuid.New(),
		}
		if createErr := db.Create(&mapping).Error; createErr != nil {
			log.Printf("demo fashion mapping ensure (non-fatal): user_id=%d err=%v", userID, createErr)
			return
		}
		log.Printf("demo fashion mapping created: user_id=%d pseudonymous_id=%s", userID, mapping.PseudonymousID)
	} else if err != nil {
		log.Printf("demo fashion mapping ensure (non-fatal): user_id=%d err=%v", userID, err)
		return
	}

	var consent consentpersistence.ConsentRow
	cerr := db.Where("user_id = ? AND sdk_version = ?", userID, demoSDKVersion).First(&consent).Error
	if cerr == gorm.ErrRecordNotFound {
		consent = consentpersistence.ConsentRow{
			UserID:       userID,
			ConsentLevel: "individual",
			SMSConsented: true,
			IsActive:     true,
			GrantedAt:    &now,
			SDKVersion:   demoSDKVersion,
		}
		if err := db.Create(&consent).Error; err != nil {
			log.Printf("demo fashion consent ensure (non-fatal): user_id=%d err=%v", userID, err)
		}
	} else if cerr != nil {
		log.Printf("demo fashion consent ensure (non-fatal): user_id=%d err=%v", userID, cerr)
	} else if !consent.SMSConsented {
		_ = db.Model(&consent).Update("sms_consented", true).Error
	}

	seedFashionIntentsForUser(db, mapping.PseudonymousID.String(), now, offset)
}

func seedDemoSMSRecipients(db *gorm.DB) {
	type demoUserMapping struct {
		UserID         int64
		PseudonymousID uuid.UUID
	}
	var demoUsers []demoUserMapping
	if err := db.Raw(`
		SELECT c.user_id, m.pseudonymous_id
		FROM consents c
		INNER JOIN pseudonymous_mappings m ON m.user_id = c.user_id
		WHERE c.sdk_version = ?
		ORDER BY c.user_id
		LIMIT ?
	`, demoSDKVersion, demoFashionCohortSize).Scan(&demoUsers).Error; err != nil {
		log.Printf("demo sms recipient seed (non-fatal): %v", err)
		return
	}

	for i := range demoUsers {
		pseudo := demoUsers[i].PseudonymousID
		userID := demoUsers[i].UserID
		updates := map[string]any{
			"phone_e164":      fmt.Sprintf("+1555000%04d", i+1),
			"display_name":    fmt.Sprintf("Demo User %02d", i+1),
			"pseudonymous_id": pseudo,
			"is_active":       true,
			"is_mock":         true,
		}

		var existing deliverypersistence.DemoSMSRecipientRow
		err := db.Where("user_id = ?", userID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			row := deliverypersistence.DemoSMSRecipientRow{
				UserID:         userID,
				PhoneE164:      updates["phone_e164"].(string),
				DisplayName:    updates["display_name"].(string),
				PseudonymousID: &pseudo,
				IsActive:       true,
				IsMock:         true,
			}
			if createErr := db.Create(&row).Error; createErr != nil {
				log.Printf("demo sms recipient seed (non-fatal): %v", createErr)
			}
			continue
		}
		if err != nil {
			log.Printf("demo sms recipient seed (non-fatal): %v", err)
			continue
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			log.Printf("demo sms recipient update (non-fatal): %v", err)
		}
	}

	var withPseudo int64
	_ = db.Model(&deliverypersistence.DemoSMSRecipientRow{}).
		Where("pseudonymous_id IS NOT NULL AND is_active = TRUE").
		Count(&withPseudo).Error
	log.Printf("demo sms recipients ready: %d with pseudonymous_id", withPseudo)
}

func seedFashionIntentsForUser(db *gorm.DB, pseudonymousID string, now time.Time, offset int) {
	var existing int64
	db.Model(&intentpersistence.IntentRow{}).
		Where("pseudonymous_id = ? AND intent_name = ?", pseudonymousID, "fashion_interest").
		Count(&existing)
	if existing >= 6 {
		return
	}

	confidence := 0.82 + float64(offset%5)*0.02
	for day := 1; day <= 6; day++ {
		createdAt := now.AddDate(0, 0, -(day + offset%3)).Truncate(24 * time.Hour).Add(time.Duration(9+day) * time.Hour)
		if err := db.Create(&intentpersistence.IntentRow{
			PseudonymousID: pseudonymousID, IntentName: "fashion_interest",
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
