package database

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"strconv"
	"strings"
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
const demoPhonePrefix = "+1555000"

// demoRecipientDisplayName keeps the label aligned with the phone slot the row owns.
func demoRecipientDisplayName(phone string) string {
	slot, err := strconv.Atoi(strings.TrimPrefix(phone, demoPhonePrefix))
	if !strings.HasPrefix(phone, demoPhonePrefix) || err != nil {
		return "Demo User"
	}
	return fmt.Sprintf("Demo User %02d", slot)
}

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
	var mappingCount int64
	if err := db.Model(&consentpersistence.PseudonymousMappingRow{}).
		Where("user_id = ?", userID).
		Count(&mappingCount).Error; err != nil {
		log.Printf("demo fashion mapping ensure (non-fatal): user_id=%d err=%v", userID, err)
		return
	}
	if mappingCount == 0 {
		mapping = consentpersistence.PseudonymousMappingRow{
			UserID:         userID,
			PseudonymousID: uuid.New(),
		}
		if createErr := db.Create(&mapping).Error; createErr != nil {
			log.Printf("demo fashion mapping ensure (non-fatal): user_id=%d err=%v", userID, createErr)
			return
		}
		log.Printf("demo fashion mapping created: user_id=%d pseudonymous_id=%s", userID, mapping.PseudonymousID)
	} else if err := db.Where("user_id = ?", userID).First(&mapping).Error; err != nil {
		log.Printf("demo fashion mapping ensure (non-fatal): user_id=%d err=%v", userID, err)
		return
	}

	var consent consentpersistence.ConsentRow
	var consentCount int64
	if err := db.Model(&consentpersistence.ConsentRow{}).
		Where("user_id = ? AND sdk_version = ?", userID, demoSDKVersion).
		Count(&consentCount).Error; err != nil {
		log.Printf("demo fashion consent ensure (non-fatal): user_id=%d err=%v", userID, err)
		return
	}
	if consentCount == 0 {
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
	} else if err := db.Where("user_id = ? AND sdk_version = ?", userID, demoSDKVersion).
		First(&consent).Error; err != nil {
		log.Printf("demo fashion consent ensure (non-fatal): user_id=%d err=%v", userID, err)
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
	// DISTINCT ON: a user with more than one demo consent would otherwise appear twice
	// and shift every following phone slot.
	if err := db.Raw(`
		SELECT DISTINCT ON (c.user_id) c.user_id, m.pseudonymous_id
		FROM consents c
		INNER JOIN pseudonymous_mappings m ON m.user_id = c.user_id
		WHERE c.sdk_version = ?
		ORDER BY c.user_id
		LIMIT ?
	`, demoSDKVersion, demoFashionCohortSize).Scan(&demoUsers).Error; err != nil {
		log.Printf("demo sms recipient seed (non-fatal): %v", err)
		return
	}

	var existingRows []deliverypersistence.DemoSMSRecipientRow
	if err := db.Find(&existingRows).Error; err != nil {
		log.Printf("demo sms recipient seed (non-fatal): %v", err)
		return
	}

	// A recipient keeps the phone number it was created with. Demo user ids are random,
	// so cohort ordering changes between boots; re-deriving the phone from the loop index
	// hands an already-claimed number to a different user and trips the unique index.
	existingByUserID := make(map[int64]deliverypersistence.DemoSMSRecipientRow, len(existingRows))
	claimedPhones := make(map[string]struct{}, len(existingRows))
	for _, row := range existingRows {
		existingByUserID[row.UserID] = row
		claimedPhones[row.PhoneE164] = struct{}{}
	}

	nextSlot := 1
	allocatePhone := func() string {
		for {
			phone := fmt.Sprintf("%s%04d", demoPhonePrefix, nextSlot)
			nextSlot++
			if _, taken := claimedPhones[phone]; !taken {
				claimedPhones[phone] = struct{}{}
				return phone
			}
		}
	}

	for i := range demoUsers {
		pseudo := demoUsers[i].PseudonymousID
		userID := demoUsers[i].UserID

		if existing, ok := existingByUserID[userID]; ok {
			updates := map[string]any{
				"pseudonymous_id": pseudo,
				"is_active":       true,
				"is_mock":         true,
			}
			if existing.DisplayName == "" {
				updates["display_name"] = demoRecipientDisplayName(existing.PhoneE164)
			}
			if err := db.Model(&deliverypersistence.DemoSMSRecipientRow{}).
				Where("user_id = ?", userID).
				Updates(updates).Error; err != nil {
				log.Printf("demo sms recipient update (non-fatal): %v", err)
			}
			continue
		}

		phone := allocatePhone()
		row := deliverypersistence.DemoSMSRecipientRow{
			UserID:         userID,
			PhoneE164:      phone,
			DisplayName:    demoRecipientDisplayName(phone),
			PseudonymousID: &pseudo,
			IsActive:       true,
			IsMock:         true,
		}
		if err := db.Create(&row).Error; err != nil {
			log.Printf("demo sms recipient seed (non-fatal): %v", err)
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
