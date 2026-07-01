package domain

import "time"

type Advertiser struct {
	ID          string
	CompanyName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
