package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis() *redis.ClusterClient {
	addrs := []string{
		"redis-node-1:6379",
		"redis-node-2:6379",
		"redis-node-3:6379",
		"redis-node-4:6379",
		"redis-node-5:6379",
		"redis-node-6:6379",
	}

	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          addrs,
		ReadOnly:       true,
		RouteByLatency: true,
	})
	ctx := context.Background()
	var err error
	for i := 0; i < 10; i++ {
		err = rdb.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
			return client.Ping(ctx).Err()
		})
		if err == nil {
			log.Println("Redis-кластер успешно подключен!")
			return rdb
		}
		time.Sleep(2 * time.Second)
	}

	return nil
}
