package domain

import (
	"time"

	"github.com/google/uuid"
)

type Developer struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
