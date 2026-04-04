package ory

import (
	"context"
	"fmt"
	"time"

	ory "github.com/ory/client-go"
	"github.com/resoul/studio.go.api/internal/config"
	"github.com/resoul/studio.go.api/internal/domain"
)

type kratosRepository struct {
	client *ory.APIClient
}

func NewKratosRepository(cfg *config.Config) domain.UserRepository {
	configuration := ory.NewConfiguration()
	configuration.Servers = ory.ServerConfigurations{
		{
			URL: cfg.Kratos.AdminURL,
		},
	}
	client := ory.NewAPIClient(configuration)
	return &kratosRepository{client: client}
}

func (r *kratosRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	identity, _, err := r.client.IdentityAPI.GetIdentity(ctx, id).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get identity from kratos: %w", err)
	}

	traits := identity.Traits.(map[string]interface{})
	email, _ := traits["email"].(string)

	isVerified := false
	for _, addr := range identity.VerifiableAddresses {
		if addr.Verified {
			isVerified = true
			break
		}
	}

	createdAt := time.Time{}
	if identity.CreatedAt != nil {
		createdAt = *identity.CreatedAt
	}
	updatedAt := time.Time{}
	if identity.UpdatedAt != nil {
		updatedAt = *identity.UpdatedAt
	}

	return &domain.User{
		ID:        identity.Id,
		Email:     email,
		Verified:  isVerified,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *kratosRepository) GetIdentity(ctx context.Context, id string) (*domain.User, error) {
	return r.FindByID(ctx, id)
}
