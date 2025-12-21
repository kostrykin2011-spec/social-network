package service

import (
	"fmt"
	"social-network/pkg/models"
	"social-network/pkg/repository"

	"github.com/google/uuid"
)

type ProfileService interface {
	GetById(userId uuid.UUID) (*models.Profile, error)
	SearchProfile(firstName, lastName string, limit, offset int) ([]*models.Profile, error)
}

type profileService struct {
	repository repository.ProfileRepository
}

func InitProfileService(profileRepository repository.ProfileRepository) ProfileService {
	return &profileService{repository: profileRepository}
}

func (service *profileService) GetById(userId uuid.UUID) (*models.Profile, error) {
	return service.repository.GetByUserId(userId)
}
func (service *profileService) SearchProfile(firstName, lastName string, limit, offset int) ([]*models.Profile, error) {
	if firstName == "" || lastName == "" {
		return nil, fmt.Errorf("Не переданы обязательные параметры")
	}

	return service.repository.SearchProfiles(firstName, lastName, limit, offset)
}
