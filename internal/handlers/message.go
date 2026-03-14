package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"social-network/pkg/pb/chat"
	"strconv"

	"social-network/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/metadata"
)

type MessageRequest struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	SenderId    uuid.UUID `json:"sender_id"`
	RecipientId uuid.UUID `json:"recipient_id"`
	Content     string    `json:"content"`
}

type ChatHandler interface {
	Send(w http.ResponseWriter, r *http.Request)
	GetMessages(w http.ResponseWriter, r *http.Request)
}

type chatHandler struct {
	service chat.ChatServiceClient
	log     *logger.Logger
}

func InitChatHandler(service chat.ChatServiceClient, log *logger.Logger) ChatHandler {
	return &chatHandler{
		service: service,
		log:     log,
	}
}

func (handler *chatHandler) Send(w http.ResponseWriter, r *http.Request) {
	requestID := logger.GetRequestID(r.Context())
	log := handler.getLogger(r.Context())

	senderUserID, ok := r.Context().Value("user_id").(string)
	if !ok || senderUserID == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("ID отправителя не найден в контексте")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	recipientID := vars["user_id"]

	if recipientID == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("ID получателя не найден в URL")
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(recipientID)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("ID получателя в URL невалидный")
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Ошибка парсинга запроса")
		http.Error(w, "Не валидный запрос", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		log.Error().Err(err).Msg("Не передан текст сообщения")
		http.Error(w, "Не передан текст сообщения", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("sender_id", senderUserID).
		Str("recipient_id", recipientID).
		Msg("Отправка сообщения в микросервис")

	ctx := metadata.AppendToOutgoingContext(
		r.Context(),
		"x-request-id", requestID, // для логирования
		"x-user-id", senderUserID, // для аутентификации
	)

	_, err = handler.service.SendMessage(ctx, &chat.SendMessageRequest{
		RecipientId: recipientID,
		Content:     req.Content,
	})

	if err != nil {
		log.Error().Err(err).Msg("Не удалось отправить сообщение на микросервис через gRPC")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Не удалось отправить сообщение на микросервис через gRPC",
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Сообщение успешно отправлено",
	})
}

func (handler *chatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	requestID := logger.GetRequestID(r.Context())
	log := handler.getLogger(r.Context())

	currentUserID, ok := r.Context().Value("user_id").(string)
	if !ok || currentUserID == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("ID авторизованного пользователя не найден в контексте")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	recipientID := vars["user_id"]

	if recipientID == "" {
		log.Error().
			Str("request_id", requestID).
			Str("current_user", currentUserID).
			Msg("ID получателя не найден в URL")
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	limit := 50
	offset := 0

	if limitQuery := r.URL.Query().Get("limit"); limitQuery != "" {
		if value, err := strconv.Atoi(limitQuery); err == nil && value > 0 {
			limit = value
		} else {
			log.Warn().
				Str("request_id", requestID).
				Str("limit", limitQuery).
				Msg("Неверный формат limit, используется значение по умолчанию")
		}
	}

	if offsetQuery := r.URL.Query().Get("offset"); offsetQuery != "" {
		if value, err := strconv.Atoi(offsetQuery); err == nil && value >= 0 {
			offset = value
		} else {
			log.Warn().
				Str("request_id", requestID).
				Str("offset", offsetQuery).
				Msg("Неверный формат offset, используется значение по умолчанию")
		}
	}

	log.Info().
		Str("request_id", requestID).
		Str("current_user", currentUserID).
		Str("recipient_id", recipientID).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Запрос списка сообщений")

	ctx := metadata.AppendToOutgoingContext(
		r.Context(),
		"x-request-id", requestID,
		"x-user-id", currentUserID,
	)

	// Получаем сообщения с помощью gRPC-метода GetMessagesByUser из микросервиса чатов
	response, err := handler.service.GetMessagesByUser(ctx, &chat.GetDialogRequest{
		FriendId: recipientID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("current_user", currentUserID).
			Str("recipient_id", recipientID).
			Msg("Ошибка получения сообщений из микросервиса")

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Не удалось получить сообщения из микросервиса",
		})

		return
	}

	messages := make([]map[string]interface{}, 0, len(response.Messages))
	for _, msg := range response.Messages {
		messages = append(messages, map[string]interface{}{
			"sender_id":    msg.SenderId,
			"recipient_id": msg.RecipientId,
			"content":      msg.Content,
		})
	}

	log.Info().
		Str("request_id", requestID).
		Str("current_user", currentUserID).
		Str("recipient_id", recipientID).
		Int("count", len(messages)).
		Msg("Сообщения успешно получены")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

func (s *chatHandler) getLogger(ctx context.Context) *logger.Logger {
	requestID := logger.GetRequestID(ctx)
	return s.log.WithRequestID(requestID)
}
