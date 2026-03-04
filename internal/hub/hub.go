package hub

import (
	"context"
	"encoding/json"
	"log"
	"social-network/internal/queue"
	"social-network/pkg/models"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *Hub
	closeOnce sync.Once
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.Send)
	})
}

type Hub struct {
	clients       map[*Client]bool
	clientDevices map[uuid.UUID][]*Client
	rabbitMQ      *queue.RabbitMQClient
	queueName     string
	serverID      string
	register      chan *Client
	unregister    chan *Client
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

func InitWebsocketHub(rabbitMQ *queue.RabbitMQClient, serverID string) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	hub := &Hub{
		clients:       make(map[*Client]bool),
		clientDevices: make(map[uuid.UUID][]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		rabbitMQ:      rabbitMQ,
		serverID:      serverID,
		ctx:           ctx,
		cancel:        cancel,
	}

	if rabbitMQ != nil {
		var err error
		hub.queueName, err = rabbitMQ.CreateTemporaryQueue(serverID)
		if err != nil {
			log.Printf("Не удалось создать временную очередь: %v", err)
		} else {
			go hub.consumeEvents()
		}
	}

	return hub
}

// Потребляем события из RabbitMQ и доставляем клиентам
func (h *Hub) consumeEvents() {
	if h.rabbitMQ == nil || h.queueName == "" {
		return
	}

	consumerTag := "hub-" + h.serverID
	err := h.rabbitMQ.ConsumeMessagesToWS(h.ctx, h.queueName, consumerTag, h.SendToUser)

	if err != nil {
		log.Printf("Не удалось подписаться на события о публикации постов: %v", err)
	}
}

func (h *Hub) Run() {
	for {
		select {
		case <-h.ctx.Done():
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.clientDevices[client.UserID] = append(h.clientDevices[client.UserID], client)
			h.mu.Unlock()

			// Привязываем пользователя к очереди сервера
			err := h.rabbitMQ.BindQueueToUser(h.queueName, client.UserID)
			if err == nil {
				return
			}
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)

				clients := h.clientDevices[client.UserID]
				for i, c := range clients {
					if c == client {
						h.clientDevices[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}

				// Если у пользователя больше нет соединений, отвязываем от очереди
				if len(h.clientDevices[client.UserID]) == 0 {
					delete(h.clientDevices, client.UserID)
					if h.rabbitMQ != nil && h.queueName != "" {
						go func(userID uuid.UUID) {
							if err := h.rabbitMQ.UnbindQueueFromUser(h.queueName, userID); err != nil {
								log.Printf("Не удалось отвязать пользователя %s от очереди %v", userID, err)
							}
						}(client.UserID)
					}
				}

				client.Close()
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register() chan<- *Client {
	return h.register
}

func (h *Hub) Unregister() chan<- *Client {
	return h.unregister
}

func (h *Hub) SendToUser(userID uuid.UUID, message *models.WSPostFeedEvent) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.mu.RLock()
	devices := h.clientDevices[userID]
	count := len(devices)
	h.mu.RUnlock()

	if count == 0 {
		log.Printf("Нет онлайн устройств для пользователя %s", userID)
		return nil
	}

	for _, device := range devices {
		select {
		case device.Send <- data:
		default:
		}
	}

	return nil
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Hub.Unregister() <- c
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) Stop() {
	h.cancel()
}
