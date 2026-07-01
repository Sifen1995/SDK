package domain

import (
	"time"

	"github.com/google/uuid"
)

type APIKey struct {
	ID             uuid.UUID
	ApplicationID  uuid.UUID
	KeyValue       string
	SecretKeyValue string
	IsActive       bool
	RateLimit      int
	CreatedAt      time.Time
	ExpiresAt      *time.Time
}
