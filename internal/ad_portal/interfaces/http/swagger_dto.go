package http

type PortalUserDTO struct {
	ID           string `json:"id" example:"767b4966-d46b-4b11-ac84-cff55d4ab780"`
	Email        string `json:"email" example:"analyst@example.com"`
	Name         string `json:"name" example:"Sam Analyst"`
	Role         string `json:"role" example:"read_only_analyst"`
	RoleID       string `json:"role_id" example:"14dd2ce5-b97b-4443-868d-1d58add26423"`
	AdvertiserID string `json:"advertiser_id,omitempty" example:""`
	AnalystID    string `json:"analyst_id,omitempty" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	CompanyName  string `json:"company_name,omitempty" example:""`
	IsActive     bool   `json:"is_active" example:"true"`
}

type AdvertiserUserDTO struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email        string `json:"email" example:"ops@kaldi.test"`
	Name         string `json:"name" example:"Kaldi Ops"`
	Role         string `json:"role" example:"advertiser"`
	RoleID       string `json:"role_id" example:"660e8400-e29b-41d4-a716-446655440001"`
	AdvertiserID string `json:"advertiser_id" example:"770e8400-e29b-41d4-a716-446655440002"`
	CompanyName  string `json:"company_name" example:"Kaldi Coffee"`
	IsActive     bool   `json:"is_active" example:"true"`
}

type RegisterResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"ops@kaldi.test"`
	Name      string `json:"name" example:"Kaldi Ops"`
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

type CreateUserResponse struct {
	User PortalUserDTO `json:"user"`
}
