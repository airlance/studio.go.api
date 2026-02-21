package http

import (
	"github.com/football.manager.api/internal/usecase"
)

func toRegisterDTO(req RegisterRequest) usecase.RegisterDTO {
	return usecase.RegisterDTO{
		Email:    req.Email,
		Password: req.Password,
	}
}

func toVerifyEmailDTO(req VerifyEmailRequest) usecase.VerifyEmailDTO {
	return usecase.VerifyEmailDTO{
		Email: req.Email,
		Code:  req.Code,
	}
}

func toResendVerificationDTO(req ResetPasswordRequest) usecase.ResendVerificationDTO {
	return usecase.ResendVerificationDTO{
		Email: req.Email,
	}
}

func toLoginDTO(req LoginRequest) usecase.LoginDTO {
	return usecase.LoginDTO{
		Email:    req.Email,
		Password: req.Password,
	}
}

func toResetPasswordRequestDTO(req ResetPasswordRequest) usecase.ResetPasswordRequestDTO {
	return usecase.ResetPasswordRequestDTO{
		Email: req.Email,
	}
}

func toResetPasswordDTO(req ConfirmResetPasswordRequest) usecase.ResetPasswordDTO {
	return usecase.ResetPasswordDTO{
		Email:       req.Email,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	}
}

func toCreateManagerDTO(req CreateManagerRequest) usecase.CreateManagerDTO {
	return usecase.CreateManagerDTO{
		Name:      req.Name,
		Status:    req.Status,
		CountryID: req.CountryID,
		Avatar:    req.Avatar,
	}
}

func toCreateCareerDTO(req CreateCareerRequest) usecase.CreateCareerDTO {
	return usecase.CreateCareerDTO{
		Name: req.Name,
	}
}

func toUserResponse(dto *usecase.UserDTO) UserResponse {
	if dto == nil {
		return UserResponse{}
	}

	return UserResponse{
		ID:              dto.UUID,
		Email:           dto.Email,
		EmailVerified:   dto.EmailVerified,
		EmailVerifiedAt: dto.EmailVerifiedAt,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
	}
}

func toCountryResponse(dto *usecase.CountryDTO) *CountryResponse {
	if dto == nil {
		return nil
	}

	return &CountryResponse{
		ID:   dto.ID,
		Code: dto.Code,
		Name: dto.Name,
	}
}

func toManagerResponse(dto *usecase.ManagerDTO) ManagerResponse {
	if dto == nil {
		return ManagerResponse{}
	}

	return ManagerResponse{
		ID:        dto.ID,
		UserID:    dto.UserID,
		Name:      dto.Name,
		Status:    dto.Status,
		CountryID: dto.CountryID,
		Avatar:    dto.Avatar,
		Country:   toCountryResponse(dto.Country),
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

func toCareerResponse(dto *usecase.CareerDTO) CareerResponse {
	if dto == nil {
		return CareerResponse{}
	}

	return CareerResponse{
		ID:        dto.ID,
		ManagerID: dto.ManagerID,
		Name:      dto.Name,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}
