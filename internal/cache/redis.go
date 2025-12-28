package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"social-network/internal/config"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

func InitRedis(config *config.Config) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisConfig.RedisHost, config.RedisConfig.RedisPort),
		Password: config.RedisConfig.RedisPassword,
		DB:       config.RedisConfig.RedisDB,
		PoolSize: config.RedisConfig.RedisPoolSize,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis successfully")

	return nil
}

func Close() error {
	return redisClient.Close()
}

func GetClient() *redis.Client {
	return redisClient
}

// Ключ для ленты пользователя
func FeedKey(userID string) string {
	return fmt.Sprintf("feed:%s", userID)
}

// Ключ для поста
func PostKey(postID string) string {
	return fmt.Sprintf("post:%s", postID)
}

// Ключ для автора поста
func UserPostsKey(userID string) string {
	return fmt.Sprintf("user_posts:%s", userID)
}

// Ключ для друга
func FriendsKey(userID string) string {
	return fmt.Sprintf("friends:%s", userID)
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return redisClient.Set(ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	return redisClient.Get(ctx, key).Result()
}

// Удаления одного или нескольких ключей
func Del(keys ...string) error {
	return redisClient.Del(ctx, keys...).Err()
}

// Добавление элемента с баллом
func ZAdd(key string, members ...*redis.Z) error {
	return redisClient.ZAdd(ctx, key, members...).Err()
}

// Получение по диапозону
func ZRange(key string, start, stop int64) ([]string, error) {
	return redisClient.ZRange(ctx, key, start, stop).Result()
}

func ZRevRange(key string, start, stop int64) ([]string, error) {
	return redisClient.ZRevRange(ctx, key, start, stop).Result()
}

// Удаление элемента
func ZRem(key string, members ...interface{}) error {
	return redisClient.ZRem(ctx, key, members...).Err()
}

func ZCard(key string) (int64, error) {
	return redisClient.ZCard(ctx, key).Result()
}

func ZRemRangeByRank(key string, start, stop int64) error {
	return redisClient.ZRemRangeByRank(ctx, key, start, stop).Err()
}

func Exists(key string) (bool, error) {
	result, err := redisClient.Exists(ctx, key).Result()
	return result > 0, err
}

func Expire(key string, expiration time.Duration) error {
	return redisClient.Expire(ctx, key, expiration).Err()
}

func Pipeline() redis.Pipeliner {
	return redisClient.Pipeline()
}
