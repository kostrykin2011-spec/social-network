package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CounterCacheRepository struct {
	redisClient *redis.Client
}

// Инициализация репозитория Redis
func InitCacheCounterRepository(redisClient *redis.Client) *CounterCacheRepository {
	return &CounterCacheRepository{
		redisClient: redisClient,
	}
}

// Возвращает количество непрочитанных сообщений
func (repository *CounterCacheRepository) Get(ctx context.Context, userId, recipientId string) (int64, error) {
	key := fmt.Sprintf("%s:%s:count", userId, recipientId)
	count, err := repository.redisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}

	return count, nil
}

// Записываем полученное количество непрочитанных сообщений из БД в Redis
func (repository *CounterCacheRepository) Set(ctx context.Context, userId, recipientId string, value int64) error {
	key := fmt.Sprintf("%s:%s:count", userId, recipientId)
	return repository.redisClient.Set(ctx, key, value, 60*time.Minute).Err()
}

// Увеличиваем на 1
func (repository *CounterCacheRepository) Increment(ctx context.Context, sagaId, userId, recipientId string) error {
	countKey := fmt.Sprintf("%s:%s:count", userId, recipientId)
	processedKey := fmt.Sprintf("processed:%s", sagaId)

	_, err := repository.redisClient.Set(ctx, processedKey, "1", 24*time.Hour).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return fmt.Errorf("Не удалось подключиться к Redis: %w", err)
	}

	return repository.redisClient.Incr(ctx, countKey).Err()
}

// Уменьшаем на 1
func (repository *CounterCacheRepository) Decrement(ctx context.Context, userId, recipientId string) error {
	key := fmt.Sprintf("%s:%s:count", userId, recipientId)
	return repository.redisClient.Decr(ctx, key).Err()
}

// Сбросываем счетчик относительно пользователя и диалога
func (repository *CounterCacheRepository) Reset(ctx context.Context, userId, recipientId string) error {
	key := fmt.Sprintf("%s:%s:count", userId, recipientId)
	return repository.redisClient.Del(ctx, key).Err()
}
