package usecase

import "github.com/football.manager.api/internal/domain"

type RegisterDTO struct {
	Email    string
	Password string
}

type LoginDTO struct {
	Email    string
	Password string
}

type VerifyEmailDTO struct {
	Email string
	Code  string
}

type ResendVerificationDTO struct {
	Email string
}

type ResetPasswordRequestDTO struct {
	Email string
}

type ResetPasswordDTO struct {
	Email       string
	Code        string
	NewPassword string
}

type UserDTO struct {
	ID              uint   `json:"id"`
	UUID            string `json:"uuid"`
	Email           string `json:"email"`
	EmailVerified   bool   `json:"email_verified"`
	EmailVerifiedAt *int64 `json:"email_verified_at,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

type CountryDTO struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateManagerDTO struct {
	Name      string
	Status    string
	CountryID *uint
	Avatar    string
}

type ManagerDTO struct {
	ID        uint        `json:"id"`
	UserID    uint        `json:"user_id"`
	Name      string      `json:"name"`
	Status    string      `json:"status"`
	CountryID *uint       `json:"country_id,omitempty"`
	Avatar    string      `json:"avatar,omitempty"`
	Country   *CountryDTO `json:"country,omitempty"`
	CreatedAt int64       `json:"created_at"`
	UpdatedAt int64       `json:"updated_at"`
}

type CreateCareerDTO struct {
	Name string
}

type CareerDTO struct {
	ID        uint   `json:"id"`
	ManagerID uint   `json:"manager_id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func mapManagerToDTO(manager *domain.Manager) *ManagerDTO {
	if manager == nil {
		return nil
	}

	return &ManagerDTO{
		ID:        manager.ID,
		UserID:    manager.UserID,
		Name:      manager.Name,
		Status:    manager.Status,
		CountryID: manager.CountryID,
		Avatar:    manager.Avatar,
		Country:   mapCountryToDTO(manager.Country),
		CreatedAt: manager.CreatedAt.Unix(),
		UpdatedAt: manager.UpdatedAt.Unix(),
	}
}

func mapCareerToDTO(career *domain.Career) *CareerDTO {
	if career == nil {
		return nil
	}

	return &CareerDTO{
		ID:        career.ID,
		ManagerID: career.ManagerID,
		Name:      career.Name,
		CreatedAt: career.CreatedAt.Unix(),
		UpdatedAt: career.UpdatedAt.Unix(),
	}
}

func mapCountryToDTO(country *domain.Country) *CountryDTO {
	if country == nil {
		return nil
	}

	return &CountryDTO{
		ID:   country.ID,
		Code: country.Code,
		Name: country.Name,
	}
}
