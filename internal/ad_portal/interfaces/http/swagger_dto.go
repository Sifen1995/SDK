package http

type PortalUserDTO struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email        string `json:"email" example:"advertiser@test.com"`
	Name         string `json:"name" example:"Jane Doe"`
	Role         string `json:"role" example:"advertiser"`
	RoleID       string `json:"role_id" example:"660e8400-e29b-41d4-a716-446655440001"`
	AdvertiserID string `json:"advertiser_id" example:"770e8400-e29b-41d4-a716-446655440002"`
	CompanyName  string `json:"company_name" example:"Acme Inc"`
	IsActive     bool   `json:"is_active" example:"true"`
}

type RegisterResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"advertiser@test.com"`
	Name      string `json:"name" example:"Jane Doe"`
	Role      string `json:"role" example:"advertiser"`
	CreatedAt string `json:"created_at" example:"2026-06-01T12:00:00Z"`
}

type LoginUserInfo struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role string `json:"role" example:"advertiser"`
}

type AdPortalLoginResponse struct {
	Token        string        `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string        `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt    string        `json:"expires_at" example:"2026-06-02T12:00:00Z"`
	User         LoginUserInfo `json:"user"`
}

type MeResponse struct {
	User PortalUserDTO `json:"user"`
}
