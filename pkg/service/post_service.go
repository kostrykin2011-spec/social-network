package service

import (
	"context"
	"fmt"
	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/utils"

	"github.com/google/uuid"
)

type PostService interface {
	AddPost(userId uuid.UUID, postRequest *models.CreatePostRequest) error
	GetById(postId uuid.UUID) (*models.Post, error)
	DeletePost(postId, userId uuid.UUID) error
	GetFeed(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*models.Post, error)
	GetFeedCount(ctx context.Context, userId uuid.UUID) int64
}

type postService struct {
	postRepository       repository.PostRepository
	userRepository       repository.UserRepository
	friendShipRepository repository.FriendShipRepository
	feedService          FeedService
}

// Инициализация сервиса постов
func InitPostService(postRepository repository.PostRepository, userRepository repository.UserRepository, friendShipRepository repository.FriendShipRepository, feedService FeedService) PostService {
	return &postService{postRepository: postRepository, userRepository: userRepository, feedService: feedService}
}

// Создание поста
func (service *postService) AddPost(userId uuid.UUID, postRequest *models.CreatePostRequest) error {
	user, err := service.userRepository.GetUserById(userId)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("Пользователь не найден")
	}

	// Валидируем данные по посту
	err = utils.ValidatePostRequest(postRequest.Title, postRequest.Content)
	if err != nil {
		return err
	}

	var post models.Post = models.Post{
		Id:       uuid.New(),
		UserId:   userId,
		Title:    postRequest.Title,
		Content:  postRequest.Content,
		IsPublic: true,
	}

	err = service.postRepository.AddPost(&post)
	if err != nil {
		return err
	}

	// Добавляем пост в кеш в потоке
	go func() {
		err := service.feedService.AddPostToFeed(post.UserId, &post)
		if err != nil {
			// Логгируем ошибку добавления поста в кеш
		}
	}()

	return nil
}

// Получение поста по Id
func (service *postService) GetById(postId uuid.UUID) (*models.Post, error) {
	return service.postRepository.GetById(postId)
}

// Удаление поста
func (service *postService) DeletePost(postId, userId uuid.UUID) error {
	post, err := service.postRepository.GetById(postId)
	if err != nil {
		return err
	}

	if post == nil {
		return fmt.Errorf("Пост не найден")
	}

	if userId != post.UserId {
		return fmt.Errorf("Вы не являетесь автором поста")
	}

	err = service.postRepository.DeletePost(postId, userId)
	if err != nil {
		return err
	}

	go func() {
		err := service.feedService.DeletePostInFeeds(post.UserId, post)
		if err != nil {
			// Логгируем ошибку удаления поста
		}
	}()

	return nil
}

// Лента постов
func (service *postService) GetFeed(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*models.Post, error) {
	posts, err := service.feedService.GetFeed(ctx, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// Количество постов в ленте
func (service *postService) GetFeedCount(ctx context.Context, userId uuid.UUID) int64 {
	result := service.feedService.GetFeedCountByUser(ctx, userId)

	return result
}
