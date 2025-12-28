package service

import (
	"context"
	"fmt"
	"social-network/pkg/database"
	"social-network/pkg/models"
	"social-network/pkg/repository"

	"github.com/google/uuid"
)

type ProfileService interface {
	GetById(ctx context.Context, userId uuid.UUID) (*models.Profile, error)
	SearchProfile(ctx context.Context, firstName, lastName string, limit, offset int) ([]*models.Profile, error)
}

type profileService struct {
	repository repository.ProfileRepository
}

func InitProfileService(profileRepository repository.ProfileRepository) ProfileService {
	return &profileService{repository: profileRepository}
}

func (service *profileService) GetById(ctx context.Context, userId uuid.UUID) (*models.Profile, error) {
	ctx = database.WithReplica(ctx)

	return service.repository.GetByUserId(ctx, userId)
}
func (service *profileService) SearchProfile(ctx context.Context, firstName, lastName string, limit, offset int) ([]*models.Profile, error) {
	if firstName == "" || lastName == "" {
		return nil, fmt.Errorf("Не переданы обязательные параметры")
	}

	ctx = database.WithReplica(ctx)

	return service.repository.SearchProfiles(ctx, firstName, lastName, limit, offset)
}
