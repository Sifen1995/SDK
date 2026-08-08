package domain

import "time"

type PortalUser struct {
	ID           string
	Email        string
	PasswordHash string
	Name         string
	RoleID       string
	AdvertiserID *string
	AnalystID    *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Role         *Role
	Advertiser   *Advertiser
	Analyst      *Analyst
}

func (u *PortalUser) RoleSlug() string {
	if u.Role != nil {
		return u.Role.Slug
	}
	return ""
}

func (u *PortalUser) AccountAdvertiserID() string {
	if u.AdvertiserID != nil {
		return *u.AdvertiserID
	}
	return ""
}

func (u *PortalUser) AccountAnalystID() string {
	if u.AnalystID != nil {
		return *u.AnalystID
	}
	return ""
}
