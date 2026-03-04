package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"social-network/internal/config"
	"social-network/pkg/models"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type FeedTask struct {
	PostID    uuid.UUID `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"` // Автор поста
	CreatedAt time.Time `json:"created_at"`
}

type RabbitMQClient struct {
	config  *config.Config
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	mu      sync.RWMutex
}

// Инициализация брокера сообщений RabbitMQ
func InitRabbitMQClient(config *config.Config) (*RabbitMQClient, error) {
	var conn *amqp.Connection
	var err error
	hosts := []string{"rabbitmq-2", "rabbitmq-3", "rabbitmq-1"}

	for _, host := range hosts {
		conn, err = amqp.Dial(config.GetConnectRabbitMQString(host))
		if err == nil {
			break
		}
	}

	if conn == nil {
		return nil, fmt.Errorf("не удалось подключиться ни к одному узлу RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Не удалось открыть канал: %w", err)
	}

	err = ch.Qos(10, 0, false)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Не удалось установить ограничение для неподтвержденных сообщений: %w", err)
	}

	ch.ExchangeDeclare("feed.tasks", "topic", true, false, false, false, nil)
	ch.ExchangeDeclare("feed.events", "topic", true, false, false, false, nil)

	return &RabbitMQClient{
		config:  config,
		conn:    conn,
		channel: ch,
	}, nil
}

// Создание очереди для задач
func (r *RabbitMQClient) CreateTaskQueue(name string) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	args := amqp.Table{
		"x-queue-type": "quorum",
	}
	_, err := ch.QueueDeclare(
		name,
		true,
		false, // Не удаляем
		false,
		false,
		args,
	)

	if err != nil {
		return fmt.Errorf("Не удалось создать очередь: %w", err)
	}

	err = ch.QueueBind(
		name,
		"#", // Получаем все
		"feed.tasks",
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("Не удалось привязать очередь к Exchange: %w", err)
	}

	return nil
}

// Создает временную очередь для WebSocket сервера
func (r *RabbitMQClient) CreateTemporaryQueue(serverID string) (string, error) {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	queueName := fmt.Sprintf("ws_%s_%d", serverID, time.Now().UnixNano())

	queue, err := ch.QueueDeclare(
		queueName,
		false,
		true, // удаляем при отключении
		false,
		false,
		nil,
	)
	if err != nil {
		return "", err
	}

	return queue.Name, nil
}

// Отправка задачи в очередь RabbitMQ
func (r *RabbitMQClient) PublishFeedTask(ctx context.Context, task *FeedTask) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("Не удалось преобразовать задачу в Json: %w", err)
	}

	if ch == nil || r.conn.IsClosed() {
		// Пробуем переподключиться к другому узлу
		if err := r.Reconnect(); err != nil {
			return fmt.Errorf("соединение потеряно и не удалось переподключиться: %w", err)
		}
		ch = r.channel
	}

	err = ch.PublishWithContext(ctx,
		"feed.tasks",
		task.UserID.String(),
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Сохранять на диск
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("Не удалось опубликовать сообщение в RabbitMQ: %w", err)
	}

	return nil
}

// Переподключение на другой хост
func (r *RabbitMQClient) Reconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel != nil {
		r.channel.Close()
		r.channel = nil
	}
	if r.conn != nil && !r.conn.IsClosed() {
		r.conn.Close()
		r.conn = nil
	}

	hosts := []string{"rabbitmq-2", "rabbitmq-3", "rabbitmq-1"}

	for _, host := range hosts {
		url := r.config.GetConnectRabbitMQString(host)
		conn, err := amqp.Dial(url)
		if err != nil {
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			continue
		}

		ch.Qos(10, 0, false)
		if err := ch.ExchangeDeclare("feed.tasks", "topic", true, false, false, false, nil); err != nil {
			conn.Close()
			continue
		}

		if err := ch.ExchangeDeclare("feed.events", "topic", true, false, false, false, nil); err != nil {
			conn.Close()
			continue
		}

		r.conn = conn
		r.channel = ch

		return nil
	}

	return fmt.Errorf("не удалось подключиться ни к одному узлу")
}

// Отправка события конкретному пользователю
func (r *RabbitMQClient) PublishEventToUser(ctx context.Context, event *models.PostFeedEvent, friendUserID uuid.UUID) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	routingKey := fmt.Sprintf("user.%s", friendUserID.String())
	if err := ch.ExchangeDeclarePassive("feed.events", "topic", true, false, false, false, nil); err != nil {
		ch.ExchangeDeclare("feed.events", "topic", true, false, false, false, nil)
	}

	err = ch.PublishWithContext(ctx,
		"feed.events",
		routingKey,
		true,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp.Table{
				"event_type": event.EventType,
				"post_id":    event.PostID.String(),
				"author_id":  event.UserID.String(),
				"friend_id":  friendUserID.String(),
			},
			Timestamp: time.Now(),
		})

	return err
}

// Привязываем к очереди определенного пользователя
func (r *RabbitMQClient) BindQueueToUser(queueName string, userId uuid.UUID) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil || r.conn == nil || r.conn.IsClosed() {
		// Пробуем переподключиться
		if err := r.Reconnect(); err != nil {
			return fmt.Errorf("соединение закрыто и не удалось переподключиться: %w", err)
		}
		ch = r.channel
	}
	routingKey := fmt.Sprintf("user.%s", userId.String())
	err := ch.QueueBind(queueName, routingKey, "feed.events", false, nil)
	if err != nil {
		return err
	}

	return nil
}

// Отвязываем от очереди конкретного пользователя по RouteKey user."" для Websocket-сервера
func (r *RabbitMQClient) UnbindQueueFromUser(queueName string, userId uuid.UUID) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	routingKey := fmt.Sprintf("user.%s", userId.String())
	err := ch.QueueUnbind(queueName, routingKey, "feed.events", nil)
	if err != nil {
		return err
	}

	return nil
}

// Подписываемся на события из временной очереди для веб-сокет сервера
func (r *RabbitMQClient) ConsumeMessagesToWS(ctx context.Context, queueName, consumerTag string, handler func(userId uuid.UUID, event *models.WSPostFeedEvent) error) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	msgs, err := ch.Consume(queueName, consumerTag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				friendIdStr, ok := msg.Headers["friend_id"].(string)
				if !ok {
					msg.Ack(false)
					continue
				}

				friendID, err := uuid.Parse(friendIdStr)
				if err != nil {
					msg.Ack(false)
					continue
				}

				var event models.WSPostFeedEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					msg.Nack(false, false)
					continue
				}

				if err := handler(friendID, &event); err != nil {
					msg.Nack(false, true)
					continue
				}

				msg.Ack(false)
			}
		}
	}()

	return nil
}

// Получаем задачи из очереди
func (r *RabbitMQClient) ConsumeMessagesTask(ctx context.Context, queueName, consumerTag string, handler func(amqp.Delivery)) error {
	log.Println(queueName)
	log.Println(consumerTag)
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	msgs, err := ch.Consume(queueName, consumerTag, false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				handler(msg)
			}
		}
	}()

	return nil
}

// Получаем задачи из очереди и отправляем в handler для worker
func (r *RabbitMQClient) ConsumeFeedTasks(ctx context.Context, queueName string, handler func(context.Context, *FeedTask) error) error {
	return r.ConsumeMessagesTask(ctx, queueName, "feed-worker", func(msg amqp.Delivery) {
		var task FeedTask
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			msg.Nack(false, false)
			return
		}
		if err := handler(ctx, &task); err != nil {
			msg.Nack(false, true)
			return
		}
		msg.Ack(false)
	})
}

// Close закрывает соединение с RabbitMQ
func (r *RabbitMQClient) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQClient) GetChannel() *amqp.Channel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.channel
}

func (r *RabbitMQClient) GetConn() *amqp.Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn
}

func (r *RabbitMQClient) IsClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn == nil || r.conn.IsClosed()
}
