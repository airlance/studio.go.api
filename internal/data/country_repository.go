package data

import (
	"context"
	"fmt"

	"github.com/football.manager.api/internal/domain"
	"gorm.io/gorm"
)

type countryRepository struct {
	db *gorm.DB
}

func NewCountryRepository(db *gorm.DB) domain.CountryRepository {
	return &countryRepository{db: db}
}

func (r *countryRepository) ListAll(ctx context.Context) ([]*domain.Country, error) {
	var models []CountryModel
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list countries: %w", err)
	}

	countries := make([]*domain.Country, 0, len(models))
	for i := range models {
		countries = append(countries, toCountryDomain(&models[i]))
	}
	return countries, nil
}
