package http

type PortalUserDTO struct {
	ID            string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email         string `json:"email" example:"advertiser@test.com"`
	Name          string `json:"name" example:"Jane Doe"`
	Role          string `json:"role" example:"advertiser"`
	RoleID        string `json:"role_id" example:"660e8400-e29b-41d4-a716-446655440001"`
	AdvertiserID  string `json:"advertiser_id" example:"770e8400-e29b-41d4-a716-446655440002"`
	CompanyName   string `json:"company_name" example:"Acme Inc"`
	IsActive      bool   `json:"is_active" example:"true"`
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
