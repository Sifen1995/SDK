package http

// PortalUserDTO is a portal user in API responses.
type PortalUserDTO struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	CompanyName  string `json:"company_name" example:"Acme Inc"`
	Email        string `json:"email" example:"advertiser@test.com"`
	ContactName  string `json:"contact_name" example:"Jane Doe"`
	Role         string `json:"role" example:"advertiser"`
	APIKey       string `json:"api_key" example:"abc123..."`
	IsActive     bool   `json:"is_active" example:"true"`
}

type RegisterResponse struct {
	User PortalUserDTO `json:"user"`
}

type AdPortalLoginResponse struct {
	Token string        `json:"token"`
	User  PortalUserDTO `json:"user"`
}

type MeResponse struct {
	User PortalUserDTO `json:"user"`
}
