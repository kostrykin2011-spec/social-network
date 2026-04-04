package services

import (
	"chat-service/internal/domain"
	"chat-service/internal/events"
	"chat-service/internal/helpers"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisMessageService struct {
	redisClient *redis.ClusterClient
	scriptSHA   string
	event       *events.SendMessageEvent
}

func InitRedisMessageService(redisClient *redis.ClusterClient, event *events.SendMessageEvent) *RedisMessageService {
	scriptText := `
	local dialog_id = redis.call('GET', KEYS[1])
	if not dialog_id then
    dialog_id = ARGV[3]
	   redis.call('SET', KEYS[1], dialog_id)
	end

	local message_payload = cjson.encode({
		content = ARGV[1],
		created_at = ARGV[2]
	})

	redis.call('ZADD', KEYS[2], ARGV[2], message_payload)
	return dialog_id`

	shaScript, err := redisClient.ScriptLoad(context.Background(), scriptText).Result()
	if err != nil {
		panic(err)
	}

	return &RedisMessageService{
		redisClient: redisClient,
		scriptSHA:   shaScript,
		event:       event,
	}
}

func (service *RedisMessageService) Send(ctx context.Context, messageRequest domain.MessageRequest, senderId, recipientId uuid.UUID) error {
	user1Id, user2Id := helpers.OrderUUIDs(senderId, recipientId)
	hashKey := fmt.Sprintf("{%s:%s}", user1Id, user2Id)
	newDialogId := uuid.New().String()
	time := time.Now().UnixNano()

	dialogKey := hashKey + ":dialog"
	messageKey := hashKey + ":message"

	_, err := service.redisClient.EvalSha(ctx, service.scriptSHA, []string{dialogKey, messageKey}, messageRequest.Content, time, newDialogId).Result()
	if err != nil {
		return fmt.Errorf("Ошибка в LUA-скрипте: %w", err)
	}

	if err := service.event.PublishMessage(ctx, senderId.String(), recipientId.String()); err != nil {
		service.redisClient.ZRemRangeByScore(ctx, messageKey, fmt.Sprintf("%d", time), fmt.Sprintf("%d", time))

		return fmt.Errorf("Ошибка - очередь для микросервиса Счетчики не доступна: %w", err)
	}

	return err
}

func (service *RedisMessageService) GetMessagesByUser(ctx context.Context, user1Id, user2Id uuid.UUID, limit, offset int) ([]domain.RedisMessageResponse, error) {
	u1Id, u2Id := helpers.OrderUUIDs(user1Id, user2Id)
	hashKey := fmt.Sprintf("{%s:%s}:message", u1Id, u2Id)
	messages, err := service.redisClient.ZRangeArgs(ctx, redis.ZRangeArgs{
		Key:   hashKey,
		Start: int64(offset),
		Stop:  int64(offset + limit - 1),
		Rev:   true,
	}).Result()

	var result []domain.RedisMessageResponse
	for _, item := range messages {
		var message domain.RedisMessageResponse
		json.Unmarshal([]byte(item), &message)

		result = append(result, message)
	}

	return result, err
}
