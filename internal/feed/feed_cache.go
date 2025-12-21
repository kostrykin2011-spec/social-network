package feed

import (
	"context"
	"encoding/json"
	"social-network/internal/cache"
	"social-network/pkg/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	feedSize            = 1000           // Размер ленты
	feedExpiration      = 24 * time.Hour // Лента хранится 24 часа
	postCacheExpiration = 24 * time.Hour // Посты храним 24 часа
)

type FeedCache struct {
}

func NewFeedCache() *FeedCache {
	return &FeedCache{}
}

// Добавление поста в ленту определенного пользователя userId
func (feed *FeedCache) AddPostIntoUserFeed(userId uuid.UUID, post *models.Post) error {
	postKey := cache.PostKey(post.Id.String())
	authorKey := cache.UserPostsKey(post.UserId.String())
	feedKey := cache.FeedKey(userId.String())
	postJSON, err := json.Marshal(post)
	if err != nil {
		return err
	}

	pipe := cache.GetClient().Pipeline()

	pipe.ZAdd(context.Background(), feedKey, &redis.Z{
		Score:  float64(post.CreatedAt.UnixNano()),
		Member: post.Id.String(),
	})

	pipe.ZRemRangeByRank(context.Background(), feedKey, 0, -feedSize-1)
	pipe.Expire(context.Background(), feedKey, feedExpiration)
	pipe.Set(context.Background(), postKey, postJSON, postCacheExpiration)

	pipe.ZAdd(context.Background(), authorKey, &redis.Z{
		Score:  float64(post.CreatedAt.UnixNano()),
		Member: post.Id.String(),
	})

	_, err = pipe.Exec(context.Background())

	return err
}

// Удаление поста из ленты пользователя
func (feed *FeedCache) DeletePostFromFeed(userId uuid.UUID, post *models.Post) error {
	feedKey := cache.FeedKey(userId.String())

	return cache.ZRem(feedKey, post.Id.String())
}

// Возвращаем ленту пользователя из кеша
func (feed *FeedCache) GetFeedByUserId(userId uuid.UUID, limit, offset int) ([]*models.Post, error) {
	feedUserKey := cache.FeedKey(userId.String())
	// Получаем все id постов в ленте конкретного пользователя
	postIds, err := cache.ZRevRange(feedUserKey, int64(offset), int64(offset+limit-1))
	if err != nil {
		return nil, err
	}

	if len(postIds) <= 0 {
		return nil, nil
	}

	var posts []*models.Post
	for _, postId := range postIds {
		postKey := cache.PostKey(postId)
		postJSON, err := cache.Get(postKey)
		if err != nil {
			if err == redis.Nil {
				continue
			}

			return nil, err
		}

		var post models.Post
		if err := json.Unmarshal([]byte(postJSON), &post); err != nil {
			return nil, err
		}

		posts = append(posts, &post)
	}

	return posts, nil
}

// Добавление поста в ленту друзей
func (feed *FeedCache) AddPostToFriendFeeds(userId uuid.UUID, post *models.Post, friendIds []uuid.UUID) error {
	if len(friendIds) == 0 {
		return nil
	}

	pipe := cache.GetClient().Pipeline()
	postKey := cache.PostKey(post.Id.String())

	postJSON, err := json.Marshal(post)
	if err != nil {
		return err
	}

	pipe.Set(context.Background(), postKey, postJSON, postCacheExpiration)

	// Добавляем пост в ленту каждого друга
	for _, friendID := range friendIds {
		feedKey := cache.FeedKey(friendID.String())
		pipe.ZAdd(context.Background(), feedKey, &redis.Z{
			Score:  float64(post.CreatedAt.UnixNano()),
			Member: post.Id.String(),
		})
		pipe.ZRemRangeByRank(context.Background(), feedKey, 0, -feedSize-1)
		pipe.Expire(context.Background(), feedKey, feedExpiration)
	}

	authorPostKey := cache.UserPostsKey(post.UserId.String())

	pipe.ZAdd(context.Background(), authorPostKey, &redis.Z{
		Score:  float64(post.CreatedAt.UnixNano()),
		Member: post.Id.String(),
	})

	_, err = pipe.Exec(context.Background())

	return err
}

// Обновление собственной ленты при добавлении/удалении друга
func (feed *FeedCache) UpdateUserFeedByAddedFriend(userId uuid.UUID, friendId uuid.UUID, isFriend bool) error {
	friendFeedKey := cache.FeedKey(friendId.String())
	postIds, err := cache.ZRange(friendFeedKey, 0, -1)
	if err != nil {
		return err
	}
	if len(postIds) <= 0 {
		return nil
	}

	pipe := cache.GetClient().Pipeline()
	feedKey := cache.FeedKey(userId.String())
	if isFriend {
		for _, postId := range postIds {
			postKey := cache.PostKey(postId)
			postJSON, err := cache.GetClient().Get(context.Background(), postKey).Result()
			if err != nil && err != redis.Nil {
				continue
			}

			if postJSON == "" {
				continue
			}

			var post models.Post
			err = json.Unmarshal([]byte(postJSON), &post)
			if err != nil {
				continue
			}

			pipe.ZAdd(context.Background(), feedKey, &redis.Z{
				Score:  float64(post.CreatedAt.UnixNano()),
				Member: post.Id.String(),
			})
		}
		pipe.ZRemRangeByRank(context.Background(), feedKey, 0, -feedSize-1)
	} else {
		for _, postId := range postIds {
			pipe.ZRem(context.Background(), feedKey, postId)
		}
	}

	pipe.Expire(context.Background(), feedKey, feedExpiration)
	_, err = pipe.Exec(context.Background())

	return err
}

// Прогрев кеша
func (feed *FeedCache) WarmUpCache(userId uuid.UUID, posts []*models.Post) error {
	if len(posts) <= 0 {
		return nil
	}

	feedKey := cache.FeedKey(userId.String())
	pipe := cache.GetClient().Pipeline()

	for _, post := range posts {
		postKey := cache.PostKey(post.Id.String())
		postJSON, err := json.Marshal(post)
		if err != nil {
			continue
		}
		pipe.Set(context.Background(), postKey, postJSON, postCacheExpiration)

		// Пост в ленту пользователя
		pipe.ZAdd(context.Background(), feedKey, &redis.Z{
			Score:  float64(post.CreatedAt.UnixNano()),
			Member: post.Id.String(),
		})
	}
	pipe.ZRemRangeByRank(context.Background(), feedKey, 0, -feedSize-1)
	pipe.Expire(context.Background(), feedKey, feedExpiration)
	_, err := pipe.Exec(context.Background())

	return err
}

// Получение количества постов в ленте пользователя
func (feed *FeedCache) GetInfoByUserFeed(userId uuid.UUID) (int64, error) {
	feedKey := cache.FeedKey(userId.String())

	return cache.ZCard(feedKey)
}

func (feed *FeedCache) DeletePost(postId uuid.UUID) error {
	postKey := cache.PostKey(postId.String())
	return cache.Del(postKey)
}
