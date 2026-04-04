package events

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type SendMessageEvent struct {
	rabbitMQConn *amqp091.Connection
}

type MessageSentEvent struct {
	SagaId      string `json:"saga_id"`
	UserId      string `json:"user_id"`
	RecipientId string `json:"recipient_id"`
}

// Инициализация события для отправки сообщений в очередь
func InitSendMessageEvent(rabbitMQConn *amqp091.Connection) *SendMessageEvent {
	return &SendMessageEvent{
		rabbitMQConn: rabbitMQConn,
	}
}

// Отправка сообщения в очередь
func (event *SendMessageEvent) PublishMessage(ctx context.Context, userId, recipientId string) error {
	channel, err := event.rabbitMQConn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare("chat_events", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}

	eventMessage := MessageSentEvent{
		SagaId:      uuid.New().String(),
		UserId:      userId,
		RecipientId: recipientId,
	}
	body, _ := json.Marshal(eventMessage)

	return channel.Publish("chat_events", "message.sent", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp091.Persistent,
	})
}
