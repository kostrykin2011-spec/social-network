package service

import (
	"context"
	"social-network/internal/cache"
	"social-network/internal/feed"

	"social-network/pkg/database"
	"social-network/pkg/models"
	"social-network/pkg/repository"

	"github.com/google/uuid"
)

type FeedService interface {
	GetFeed(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*models.Post, error)
	AddPostToFeed(ctx context.Context, userId uuid.UUID, post *models.Post) error
	DeletePostInFeeds(ctx context.Context, userId uuid.UUID, post *models.Post) error
	UpdateUserFeedByAddedFriend(ctx context.Context, userId uuid.UUID, friendId uuid.UUID, isFriend bool) error
	BuildUserFeed(ctx context.Context, userId uuid.UUID, limit, offset int) error
	GetFeedCountByUser(ctx context.Context, userId uuid.UUID) int64
}

type feedService struct {
	feedCache            feed.FeedCache
	postRepository       repository.PostRepository
	friendShipRepository repository.FriendShipRepository
}

// Инициализация сервиса по работе с лентами
func InitFeedService(feedCache *feed.FeedCache, postRepository repository.PostRepository, friendShipRepository repository.FriendShipRepository) FeedService {
	return &feedService{
		feedCache:            *feedCache,
		postRepository:       postRepository,
		friendShipRepository: friendShipRepository,
	}
}

// Получение ленты постов пользователя userId
func (feedService *feedService) GetFeed(ctx context.Context, userId uuid.UUID, limit, offset int) ([]*models.Post, error) {
	ctx = database.WithReplica(ctx)
	feedKey := cache.FeedKey(userId.String())
	exists, err := cache.Exists(feedKey)
	if err != nil {
		return nil, err
	}

	// Отсутствует лента, то прогреваем кеш
	if !exists {
		err = feedService.BuildUserFeed(ctx, userId, limit, offset)
		if err != nil {
			return nil, err
		}
	}
	// Получаем посты из кеша
	posts, err := feedService.feedCache.GetFeedByUserId(userId, limit, offset)
	if err != nil {
		return nil, err
	}

	// Количество постов в ленте относительно limit и offset
	countPosts := len(posts)

	// Если постов недостаточно, то догружаем из БД
	if countPosts < limit && (countPosts+offset) < 1000 {
		friendIds, err := feedService.friendShipRepository.GetFriendsByUserId(ctx, userId)
		if err != nil {
			return posts, err
		}

		// Последние посты в базе данных по выбранному пользователю
		friendIds = append(friendIds, userId)
		dbPosts, err := feedService.postRepository.GetListByUserIds(ctx, friendIds, limit-countPosts, offset+countPosts)
		if err != nil {
			return posts, nil
		}

		// Обновляем кеш
		for _, post := range dbPosts {
			_ = feedService.feedCache.AddPostIntoUserFeed(userId, post)
		}

		posts = append(posts, dbPosts...)
	}

	return posts, nil
}

// Количество записей в ленте пользователя
func (feedService *feedService) GetFeedCountByUser(ctx context.Context, userId uuid.UUID) int64 {
	feedKey := cache.FeedKey(userId.String())
	result, err := cache.ZCard(feedKey)
	if err != nil {
		return 0
	}

	return result
}

// Обновление кеша при добавлении/удалении пользователя (список друзей)
func (feedService *feedService) UpdateUserFeedByAddedFriend(ctx context.Context, userId uuid.UUID, friendId uuid.UUID, isFriend bool) error {
	err := feedService.feedCache.UpdateUserFeedByAddedFriend(userId, friendId, isFriend)
	if err != nil {
		return err
	}

	if isFriend {
		err = feedService.feedCache.UpdateUserFeedByAddedFriend(friendId, userId, isFriend)
		if err != nil {
			return err
		}
	}

	return nil
}

// Добавление поста в ленту автора и в ленты его друзей
func (feedService *feedService) AddPostToFeed(ctx context.Context, userId uuid.UUID, post *models.Post) error {
	ctx = database.WithReplica(ctx)

	err := feedService.feedCache.AddPostIntoUserFeed(userId, post)
	if err != nil {
		return err
	}

	// Получаем список друзей
	friendIds, err := feedService.friendShipRepository.GetFriendsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	if len(friendIds) <= 0 {
		return nil
	}

	// Добавляем пост в ленты друзей
	err = feedService.feedCache.AddPostToFriendFeeds(userId, post, friendIds)
	if err != nil {
		return err
	}

	return nil
}

// Удаление поста из ленты автора и лент друзей
func (feedService *feedService) DeletePostInFeeds(ctx context.Context, userId uuid.UUID, post *models.Post) error {
	err := feedService.feedCache.DeletePostFromFeed(userId, post)
	if err != nil {
		return err
	}

	// Получаем список друзей
	ctx = database.WithReplica(ctx)
	friendIds, err := feedService.friendShipRepository.GetFriendsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	for _, friendId := range friendIds {
		err = feedService.feedCache.DeletePostFromFeed(friendId, post)
		if err != nil {
			// Логгирование данных
		}
	}

	err = feedService.feedCache.DeletePost(post.Id)
	if err != nil {
		return err
	}

	return nil
}

// Строим и кешируем ленту постов
func (feedService *feedService) BuildUserFeed(ctx context.Context, userId uuid.UUID, limit, offset int) error {
	// Получаем список друзей
	friendIds, err := feedService.friendShipRepository.GetFriendsByUserId(ctx, userId)
	if err != nil {
		return err
	}

	// Последние посты в базе данных по выбранному пользователю
	friendIds = append(friendIds, userId)
	posts, err := feedService.postRepository.GetListByUserIds(ctx, friendIds, limit, offset)
	if err != nil {
		return err
	}

	err = feedService.feedCache.WarmUpCache(userId, posts)

	return err
}
