package services

import (
	"chat-service/internal/repository"
)

type DialogService interface {
}

type dialogService struct {
	repository repository.DialogRepository
}

func InitDialogService(repository repository.DialogRepository) DialogService {
	return &dialogService{
		repository: repository,
	}
}
