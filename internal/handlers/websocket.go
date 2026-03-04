package handlers

import (
	"log"
	"net/http"
	"social-network/internal/hub"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketHandler struct {
	hub *hub.Hub
}

func NewWebSocketHandler(hub *hub.Hub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("user_id")
	if userIdStr == "" {
		http.Error(w, "Id-пользователя не найдено", http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Id-пользователя указано некорректно", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Не удалось установить подключение к веб-сокет серверу: %v", err)
		return
	}

	client := &hub.Client{
		ID:     uuid.New(),
		UserID: userId,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register() <- client

	go client.WritePump()
}
