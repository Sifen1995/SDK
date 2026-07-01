package domain

import "time"

type Role struct {
	ID          string
	Slug        string
	DisplayName string
	CreatedAt   time.Time
}
