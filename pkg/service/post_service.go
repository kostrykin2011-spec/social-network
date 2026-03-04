package service

import (
	"context"
	"fmt"
	"log"
	"social-network/internal/queue"
	"social-network/pkg/database"
	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type PostService interface {
	AddPost(ctx context.Context, userId uuid.UUID, postRequest *models.CreatePostRequest) error
	GetById(ctx context.Context, postId uuid.UUID) (*models.Post, error)
	DeletePost(ctx context.Context, postId, userId uuid.UUID) error
	GetFeed(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*models.Post, error)
	GetFeedCount(ctx context.Context, userId uuid.UUID) int64
}

type postService struct {
	postRepository       repository.PostRepository
	userRepository       repository.UserRepository
	friendShipRepository repository.FriendShipRepository
	feedService          FeedService
	rabbitMQClient       *queue.RabbitMQClient
}

// Инициализация сервиса постов
func InitPostService(postRepository repository.PostRepository, userRepository repository.UserRepository, friendShipRepository repository.FriendShipRepository, feedService FeedService, rabbitMQClient *queue.RabbitMQClient) PostService {
	return &postService{postRepository: postRepository, userRepository: userRepository, feedService: feedService, rabbitMQClient: rabbitMQClient}
}

// Создание поста
func (service *postService) AddPost(ctx context.Context, userId uuid.UUID, postRequest *models.CreatePostRequest) error {
	ctx = database.WithMaster(ctx)
	// Валидируем данные по посту
	err := utils.ValidatePostRequest(postRequest.Title, postRequest.Content)
	if err != nil {
		return err
	}

	user, err := service.userRepository.GetUserById(ctx, userId)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("Пользователь не найден")
	}

	post := models.Post{
		Id:       uuid.New(),
		UserId:   userId,
		Title:    postRequest.Title,
		Content:  postRequest.Content,
		IsPublic: true,
	}

	// Добавляем пост в БД
	err = service.postRepository.AddPost(ctx, &post)
	if err != nil {
		return err
	}

	// Отправляем в RabbitMQ
	if service.rabbitMQClient == nil {
	} else {
		task := &queue.FeedTask{
			PostID:    post.Id,
			UserID:    post.UserId,
			CreatedAt: post.CreatedAt,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := service.rabbitMQClient.PublishFeedTask(ctx, task)
		if err != nil {
			log.Printf("Не удалось опубликовать: %v", err)
		}
	}

	return nil
}

// Получение поста по Id
func (service *postService) GetById(ctx context.Context, postId uuid.UUID) (*models.Post, error) {
	ctx = database.WithMaster(ctx)
	return service.postRepository.GetById(ctx, postId)
}

// Удаление поста
func (service *postService) DeletePost(ctx context.Context, postId, userId uuid.UUID) error {
	ctx = database.WithMaster(ctx)
	post, err := service.postRepository.GetById(ctx, postId)
	if err != nil {
		return err
	}

	if post == nil {
		return fmt.Errorf("Пост не найден")
	}

	if userId != post.UserId {
		return fmt.Errorf("Вы не являетесь автором поста")
	}

	err = service.postRepository.DeletePost(ctx, postId, userId)
	if err != nil {
		return err
	}

	go func() {
		err := service.feedService.DeletePostInFeeds(ctx, post.UserId, post)
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
