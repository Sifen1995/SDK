package http

type RegisterRequest struct {
	Name        string `json:"name" binding:"required" example:"Jane Doe"`
	Email       string `json:"email" binding:"required,email" example:"advertiser@test.com"`
	Password    string `json:"password" binding:"required,min=8" example:"SecurePass1!"`
	CompanyName string `json:"company_name" binding:"required" example:"Acme Inc"`
	Role        string `json:"role" enums:"advertiser,read_only_analyst" example:"advertiser"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CreateUserRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	CompanyName string `json:"company_name"`
	Role        string `json:"role" binding:"required,oneof=advertiser read_only_analyst operator_admin"`
}
