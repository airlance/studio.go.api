package service

import (
	"context"
	"errors"
	"time"

	"fmt"
	"github.com/resoul/studio.go.api/internal/domain"
	"gorm.io/gorm"
)

type profileService struct {
	repo    domain.ProfileRepository
	storage domain.Storage
}

func NewProfileService(repo domain.ProfileRepository, storage domain.Storage) domain.ProfileService {
	return &profileService{
		repo:    repo,
		storage: storage,
	}
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	profile, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.EnsureProfileExists(ctx, userID)
		}
		return nil, err
	}

	if profile == nil {
		return nil, nil // Or handle as error
	}

	if profile.AvatarURL != "" {
		presigned, err := s.storage.GetPresignedURL(ctx, "profiles", profile.AvatarURL, time.Hour)
		if err == nil {
			profile.AvatarURL = presigned
		}
	}

	return profile, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, userID string, input domain.UpdateProfileInput) (*domain.Profile, error) {
	profile, err := s.EnsureProfileExists(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Avatar != nil {
		objectName := fmt.Sprintf("%s/avatar", userID)
		err := s.storage.Upload(ctx, "profiles", objectName, input.Avatar, input.AvatarSize, input.AvatarType)
		if err != nil {
			return nil, fmt.Errorf("failed to upload avatar: %w", err)
		}
		profile.AvatarURL = objectName
	}

	profile.FirstName = input.FirstName
	profile.LastName = input.LastName
	profile.Completed = true
	profile.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, err
	}

	// For the response, get the presigned URL
	if profile.AvatarURL != "" {
		presigned, err := s.storage.GetPresignedURL(ctx, "profiles", profile.AvatarURL, time.Hour)
		if err == nil {
			profile.AvatarURL = presigned
		}
	}

	return profile, nil
}

func (s *profileService) EnsureProfileExists(ctx context.Context, userID string) (*domain.Profile, error) {
	profile, err := s.repo.FindByID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if profile == nil {
		profile = &domain.Profile{
			ID:        userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.repo.Create(ctx, profile); err != nil {
			return nil, err
		}
	}

	return profile, nil
}
