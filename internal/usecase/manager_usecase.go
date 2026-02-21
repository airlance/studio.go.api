package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/football.manager.api/internal/domain"
)

type ManagerUseCase interface {
	Create(ctx context.Context, userID uint, dto CreateManagerDTO) (*ManagerDTO, error)
	GetByUserID(ctx context.Context, userID uint) (*ManagerDTO, error)
	ExistsByUserID(ctx context.Context, userID uint) (bool, error)
}

type managerUseCase struct {
	managerRepo domain.ManagerRepository
}

func NewManagerUseCase(managerRepo domain.ManagerRepository) ManagerUseCase {
	return &managerUseCase{managerRepo: managerRepo}
}

func (uc *managerUseCase) Create(ctx context.Context, userID uint, dto CreateManagerDTO) (*ManagerDTO, error) {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	status := strings.TrimSpace(strings.ToLower(dto.Status))
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" && status != "banned" && status != "blocked" {
		return nil, fmt.Errorf("status must be one of: active, inactive, banned, blocked")
	}

	manager := &domain.Manager{
		UserID:    userID,
		Name:      name,
		Status:    status,
		CountryID: dto.CountryID,
		Avatar:    strings.TrimSpace(dto.Avatar),
	}

	if err := uc.managerRepo.Create(ctx, manager); err != nil {
		return nil, err
	}

	return mapManagerToDTO(manager), nil
}

func (uc *managerUseCase) GetByUserID(ctx context.Context, userID uint) (*ManagerDTO, error) {
	manager, err := uc.managerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapManagerToDTO(manager), nil
}

func (uc *managerUseCase) ExistsByUserID(ctx context.Context, userID uint) (bool, error) {
	return uc.managerRepo.ExistsByUserID(ctx, userID)
}
