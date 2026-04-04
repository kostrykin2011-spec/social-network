package cache

import (
	"context"
	"counter-service/config"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

func InitRedis(config *config.Config) (*redis.Client, error) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisConfig.RedisHost, config.RedisConfig.RedisPort),
		Password: config.RedisConfig.RedisPassword,
		DB:       config.RedisConfig.RedisDB,
		PoolSize: config.RedisConfig.RedisPoolSize,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("Не удалось подключиться к Redis: %w", err)
	}

	return redisClient, nil
}

func Close() error {
	return redisClient.Close()
}
