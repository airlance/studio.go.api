package http

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type AuthSuccessResponse struct {
	Message            string       `json:"message"`
	User               UserResponse `json:"user,omitempty"`
	Token              string       `json:"token,omitempty"`
	Role               string       `json:"role,omitempty"`
	OnboardingRequired bool         `json:"onboarding_required,omitempty"`
}

type UserResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	EmailVerifiedAt *int64 `json:"email_verified_at,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

type CountryResponse struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateManagerRequest struct {
	Name      string `json:"name" binding:"required"`
	Status    string `json:"status"`
	CountryID *uint  `json:"country_id"`
	Avatar    string `json:"avatar"`
}

type ManagerResponse struct {
	ID        uint             `json:"id"`
	UserID    uint             `json:"user_id"`
	Name      string           `json:"name"`
	Status    string           `json:"status"`
	CountryID *uint            `json:"country_id,omitempty"`
	Avatar    string           `json:"avatar,omitempty"`
	Country   *CountryResponse `json:"country,omitempty"`
	CreatedAt int64            `json:"created_at"`
	UpdatedAt int64            `json:"updated_at"`
}

type CreateCareerRequest struct {
	Name string `json:"name" binding:"required"`
}

type CareerResponse struct {
	ID        uint   `json:"id"`
	ManagerID uint   `json:"manager_id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}
