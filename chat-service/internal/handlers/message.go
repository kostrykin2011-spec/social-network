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

type MessageHandler interface {
	Send(w http.ResponseWriter, r *http.Request)
	GetMessages(w http.ResponseWriter, r *http.Request)
}

type messageHandler struct {
	messageService services.MessageService
}

func InitMessageHanler(messageService services.MessageService) MessageHandler {
	return &messageHandler{
		messageService: messageService,
	}
}

func (handler *messageHandler) Send(w http.ResponseWriter, r *http.Request) {
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

	err = handler.messageService.Send(ctx, req, senderId, recipientId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Сообщение успешно отправлено"})
}

func (handler *messageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	messages, _ := handler.messageService.GetMessagesByUser(ctx, currentUserId, recipientId, limit, offset)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(messages)
}
