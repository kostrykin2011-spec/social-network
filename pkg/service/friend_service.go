package service

import (
	"context"
	"fmt"
	"social-network/pkg/database"
	"social-network/pkg/repository"

	"github.com/google/uuid"
)

type FriendShipService interface {
	AddFiend(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error
	Delete(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error
	GetFriendsByUserId(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
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
func (service *friendShipService) AddFiend(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error {
	if userId == friendId {
		return fmt.Errorf("Вы не можете добавлять себя в друзья")
	}

	ctx = database.WithReplica(ctx)
	friend, err := service.userRepository.GetUserById(ctx, friendId)
	if err != nil {
		return err
	}
	if friend == nil {
		return fmt.Errorf("Добавляемый пользователь в друзья не найден")
	}

	ctx = database.WithMaster(ctx)
	err = service.friendRepository.Add(ctx, userId, friendId)
	if err != nil {
		return err
	}

	go func() {
		err := service.feedService.UpdateUserFeedByAddedFriend(ctx, userId, friendId, true)
		if err != nil {
			// Логгируем ошибку
		}
	}()

	return nil
}

// Удаление пользователя из списка друзей
func (service *friendShipService) Delete(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error {
	ctx = database.WithMaster(ctx)
	if userId == friendId {
		return fmt.Errorf("Вы не можете удалить себя из списка друзей")
	}

	err := service.friendRepository.Delete(ctx, userId, friendId)
	if err != nil {
		return err
	}

	go func() {
		err := service.feedService.UpdateUserFeedByAddedFriend(ctx, userId, friendId, false)
		if err != nil {
			// Логгируем ошибку
		}
	}()

	return nil
}

// Список друзей
func (service *friendShipService) GetFriendsByUserId(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	ctx = database.WithReplica(ctx)
	return service.friendRepository.GetFriendsByUserId(ctx, userId)
}
