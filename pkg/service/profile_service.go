package service

import (
	"social-network/pkg/models"
	"social-network/pkg/repository"

	"github.com/google/uuid"
)

type ProfileService interface {
	GetById(userId uuid.UUID) (*models.Profile, error)
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
