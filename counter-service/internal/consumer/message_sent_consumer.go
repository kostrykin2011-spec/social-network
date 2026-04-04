package consumer

import (
	"context"
	"counter-service/internal/models"
	"counter-service/internal/repository"
	logger "counter-service/pkg"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MessageSentConsumer struct {
	rabbitMQConn    *amqp.Connection
	cacheRepository *repository.CounterCacheRepository
	dbRepository    *repository.CounterDBRepository
	log             *logger.Logger
}

func InitMessageConsumer(rabbitMQConn *amqp.Connection, cacheRepository *repository.CounterCacheRepository, dbRepository *repository.CounterDBRepository, log *logger.Logger) *MessageSentConsumer {
	return &MessageSentConsumer{
		rabbitMQConn:    rabbitMQConn,
		cacheRepository: cacheRepository,
		dbRepository:    dbRepository,
		log:             log,
	}
}

func (consumer *MessageSentConsumer) Start(ctx context.Context) error {
	ch, err := consumer.rabbitMQConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("chat_events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare("chat_events_retry_ex", "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	args := amqp.Table{
		"x-dead-letter-exchange": "chat_events_retry_ex",
	}

	queue, err := ch.QueueDeclare("counter_sent_message", true, false, false, false, args)
	if err != nil {
		return err
	}

	retryArgs := amqp.Table{
		"x-dead-letter-exchange":    "chat_events",
		"x-dead-letter-routing-key": "message.sent",
		"x-message-ttl":             int32(3600000),
	}

	_, err = ch.QueueDeclare("counter_sent_message_wait", true, false, false, false, retryArgs)
	if err != nil {
		return err
	}

	err = ch.QueueBind("counter_sent_message_wait", "", "chat_events_retry_ex", false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(queue.Name, "message.sent", "chat_events", false, nil)
	if err != nil {
		return err
	}

	messages, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	consumer.log.Info().Msg("Запущен consumer, получающий информацию о новых сообщениях")
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, ok := <-messages:
			if !ok {
				consumer.log.Error().Msg("Канал закрыт - RabbitMQ")
				return fmt.Errorf("Канал закрыт - RabbitMQ")
			}
			consumer.HandleMessageByQueue(ctx, message)
		}
	}
}

// Обработка письма, полученного из очереди
func (consumer *MessageSentConsumer) HandleMessageByQueue(ctx context.Context, message amqp.Delivery) {
	const maxRetries = 10
	if death, ok := message.Headers["x-death"].([]interface{}); ok && len(death) > 0 {
		if death, ok := death[0].(amqp.Table); ok {
			if count, ok := death["count"].(int64); ok && count >= maxRetries {
				consumer.log.Error().
					Int64("retry_count", count).
					Str("body", string(message.Body)).
					Msg("Превышено количество попыток! Сообщение удалено.")

				message.Ack(false)
				return
			}
		}
	}

	var event models.MessageSentEvent
	if err := json.Unmarshal(message.Body, &event); err != nil {
		consumer.log.Error().Err(err).Msg("JSON поврежден")
		message.Ack(false)
		return
	}

	if err := consumer.cacheRepository.Increment(ctx, event.SagaId, event.UserId, event.RecipientId); err != nil {
		consumer.log.Warn().Msg("Redis недоступен, отправляем на ретрай")
		message.Nack(false, false)
		return
	}

	if err := consumer.dbRepository.Increment(ctx, event.SagaId, event.UserId, event.RecipientId); err != nil {
		consumer.log.Warn().Msg("База данных недоступна, откат Redis и ретрай")
		consumer.cacheRepository.Decrement(ctx, event.UserId, event.RecipientId)

		message.Nack(false, false)
		return
	}

	message.Ack(false)
}
