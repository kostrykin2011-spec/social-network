package service

import (
	"fmt"
	"social-network/pkg/repository"

	"github.com/google/uuid"
)

type FriendShipService interface {
	AddFiend(userId uuid.UUID, friendId uuid.UUID) error
	Delete(userId uuid.UUID, friendId uuid.UUID) error
	GetFriendsByUserId(userId uuid.UUID) ([]uuid.UUID, error)
}

type friendShipService struct {
	userRepository   repository.UserRepository
	friendRepository repository.FriendShipRepository
	feedService      FeedService
}

// Инициализация сервиса добавления/удаления друзей
func InitFriendShipService(userRepository repository.UserRepository, friendRepository repository.FriendShipRepository, feedService FeedService) FriendShipService {
	return &friendShipService{userRepository: userRepository, friendRepository: friendRepository, feedService: feedService}
}

// Добавление пользователя в список друзей
func (service *friendShipService) AddFiend(userId uuid.UUID, friendId uuid.UUID) error {
	if userId == friendId {
		return fmt.Errorf("Вы не можете добавлять себя в друзья")
	}

	friend, err := service.userRepository.GetUserById(friendId)
	if err != nil {
		return err
	}
	if friend == nil {
		return fmt.Errorf("Добавляемый пользователь в друзья не найден")
	}

	err = service.friendRepository.Add(userId, friendId)
	if err != nil {
		return err
	}

	go func() {
		err := service.feedService.UpdateUserFeedByAddedFriend(userId, friendId, true)
		if err != nil {
			// Логгируем ошибку
		}
	}()

	return nil
}

// Удаление пользователя из списка друзей
func (service *friendShipService) Delete(userId uuid.UUID, friendId uuid.UUID) error {
	if userId == friendId {
		return fmt.Errorf("Вы не можете удалить себя из списка друзей")
	}

	err := service.friendRepository.Delete(userId, friendId)
	if err != nil {
		return err
	}

	go func() {
		err := service.feedService.UpdateUserFeedByAddedFriend(userId, friendId, false)
		if err != nil {
			// Логгируем ошибку
		}
	}()

	return nil
}

// Список друзей
func (service *friendShipService) GetFriendsByUserId(userId uuid.UUID) ([]uuid.UUID, error) {
	return service.friendRepository.GetFriendsByUserId(userId)
}
