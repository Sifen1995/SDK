package domain

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID          uuid.UUID
	DeveloperID uuid.UUID
	AppName     string
	Platform    string
	BundleID    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	APIKeys     []APIKey
}
