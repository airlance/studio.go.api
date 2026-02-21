package data

import (
	"github.com/football.manager.api/internal/domain"
)

func toUserDomain(model *UserModel) *domain.User {
	if model == nil {
		return nil
	}
	return &domain.User{
		ID:                     model.ID,
		UUID:                   model.UUID,
		Email:                  model.Email,
		PasswordHash:           model.PasswordHash,
		Role:                   model.Role,
		RegistrationIP:         model.RegistrationIP,
		RegistrationUserAgent:  model.RegistrationUserAgent,
		LoginCount:             model.LoginCount,
		EmailVerifiedAt:        model.EmailVerifiedAt,
		VerificationCode:       model.VerificationCode,
		VerificationExpiresAt:  model.VerificationExpiresAt,
		ResetPasswordCode:      model.ResetPasswordCode,
		ResetPasswordExpiresAt: model.ResetPasswordExpiresAt,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}
}

func toUserModel(entity *domain.User) *UserModel {
	if entity == nil {
		return nil
	}
	return &UserModel{
		ID:                     entity.ID,
		UUID:                   entity.UUID,
		Email:                  entity.Email,
		PasswordHash:           entity.PasswordHash,
		Role:                   entity.Role,
		RegistrationIP:         entity.RegistrationIP,
		RegistrationUserAgent:  entity.RegistrationUserAgent,
		LoginCount:             entity.LoginCount,
		EmailVerifiedAt:        entity.EmailVerifiedAt,
		VerificationCode:       entity.VerificationCode,
		VerificationExpiresAt:  entity.VerificationExpiresAt,
		ResetPasswordCode:      entity.ResetPasswordCode,
		ResetPasswordExpiresAt: entity.ResetPasswordExpiresAt,
		CreatedAt:              entity.CreatedAt,
		UpdatedAt:              entity.UpdatedAt,
	}
}

func toCountryDomain(model *CountryModel) *domain.Country {
	if model == nil {
		return nil
	}
	return &domain.Country{
		ID:        model.ID,
		Code:      model.Code,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toCountryModel(entity *domain.Country) *CountryModel {
	if entity == nil {
		return nil
	}
	return &CountryModel{
		ID:        entity.ID,
		Code:      entity.Code,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func toManagerDomain(model *ManagerModel) *domain.Manager {
	if model == nil {
		return nil
	}
	return &domain.Manager{
		ID:        model.ID,
		UserID:    model.UserID,
		Name:      model.Name,
		Status:    model.Status,
		CountryID: model.CountryID,
		Avatar:    model.Avatar,
		Country:   toCountryDomain(model.Country),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toManagerModel(entity *domain.Manager) *ManagerModel {
	if entity == nil {
		return nil
	}
	return &ManagerModel{
		ID:        entity.ID,
		UserID:    entity.UserID,
		Name:      entity.Name,
		Status:    entity.Status,
		CountryID: entity.CountryID,
		Avatar:    entity.Avatar,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}

func toCareerDomain(model *CareerModel) *domain.Career {
	if model == nil {
		return nil
	}
	return &domain.Career{
		ID:        model.ID,
		ManagerID: model.ManagerID,
		Name:      model.Name,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toCareerModel(entity *domain.Career) *CareerModel {
	if entity == nil {
		return nil
	}
	return &CareerModel{
		ID:        entity.ID,
		ManagerID: entity.ManagerID,
		Name:      entity.Name,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
