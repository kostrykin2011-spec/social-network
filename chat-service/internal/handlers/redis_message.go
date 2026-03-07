package handlers

import (
	"chat-service/internal/domain"
	"chat-service/internal/services"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RedisMessageHandler interface {
	SendMessage(w http.ResponseWriter, r *http.Request)
	GetMessagesByUser(w http.ResponseWriter, r *http.Request)
}

type redisMessageHandler struct {
	service *services.RedisMessageService
}

func InitRedisMessageHanler(service *services.RedisMessageService) RedisMessageHandler {
	return &redisMessageHandler{
		service: service,
	}
}

func (handler *redisMessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//senderId, _ := helpers.Instance().GetRandom()
	senderId, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Идентификатор авторизованного пользователя указан некорректно", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	recipientId, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Идентификатор пользователя, которому отправляется сообщение, указан некорректно", http.StatusBadRequest)
		return
	}
	var req domain.MessageRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = handler.service.Send(ctx, req, senderId, recipientId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Сообщение успешно отправлено"})
}

func (handler *redisMessageHandler) GetMessagesByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//currentUserId, _ := helpers.Instance().GetRandom()
	currentUserId, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Идентификатор авторизованного пользователя указан некорректно", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	recipientId, err := uuid.Parse(vars["user_id"])
	if err != nil {
		http.Error(w, "Идентификатор пользователя, которому отправляется сообщение, указан некорректно", http.StatusBadRequest)
		return
	}

	limit := 50
	offset := 0

	if limitQuery := r.URL.Query().Get("limit"); limitQuery != "" {
		if value, err := strconv.Atoi(limitQuery); err == nil && value > 0 {
			limit = value
		}
	}

	if offsetQuery := r.URL.Query().Get("offset"); offsetQuery != "" {
		if value, err := strconv.Atoi(offsetQuery); err == nil && value >= 0 {
			offset = value
		}
	}

	messages, _ := handler.service.GetMessagesByUser(ctx, currentUserId, recipientId, limit, offset)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(messages)
}
