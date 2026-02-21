package usecase

import (
	"context"

	"github.com/football.manager.api/internal/domain"
)

type CountryUseCase interface {
	ListAll(ctx context.Context) ([]*CountryDTO, error)
}

type countryUseCase struct {
	countryRepo domain.CountryRepository
}

func NewCountryUseCase(countryRepo domain.CountryRepository) CountryUseCase {
	return &countryUseCase{countryRepo: countryRepo}
}

func (uc *countryUseCase) ListAll(ctx context.Context) ([]*CountryDTO, error) {
	countries, err := uc.countryRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]*CountryDTO, 0, len(countries))
	for _, country := range countries {
		dtos = append(dtos, mapCountryToDTO(country))
	}
	return dtos, nil
}
