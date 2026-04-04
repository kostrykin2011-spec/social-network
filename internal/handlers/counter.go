package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"social-network/pkg/pb/counter"

	"social-network/pkg/logger"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/metadata"
)

type CounterCountResponse struct {
	UserId   string `json:"user_id"`
	DialogId string `json:"dialog_id"`
	Count    int64  `json:"count"`
}

type CounterResetResponse struct {
	UserId   string `json:"user_id"`
	DialogId string `json:"dialog_id"`
	Result   string `json:"result"`
}

type CounterMessageHandler struct {
	service counter.CounterServiceClient
	log     *logger.Logger
}

func InitCounterHandler(service counter.CounterServiceClient, log *logger.Logger) *CounterMessageHandler {
	return &CounterMessageHandler{
		service: service,
		log:     log,
	}
}

func (handler *CounterMessageHandler) GetCountMessagesByDialog(w http.ResponseWriter, r *http.Request) {
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
	userId := vars["user_id"]

	if userId == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("Id пользователя не найден")
		http.Error(w, "userId не найден", http.StatusBadRequest)
		return
	}

	if senderUserID == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("Id диалога не найден")
		http.Error(w, "Id текущего пользователя не найден", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(userId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Id пользователя невалидный")
		http.Error(w, "Id пользователя невалидный", http.StatusBadRequest)
		return
	}

	ctx := metadata.AppendToOutgoingContext(
		r.Context(),
		"x-request-id", requestID, // для логирования
		"x-user-id", senderUserID, // для аутентификации
	)

	counterResponse, err := handler.service.GetCount(ctx, &counter.GetCountRequest{
		RecipientId: userId,
	})

	if err != nil {
		log.Error().Err(err).Msg("Не удалось получить данные микросервиса Счетчики через gRPC")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Не удалось получить данные микросервиса Счетчики через gRPC",
		})
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId).
		Str("sender_id", senderUserID).
		Int64("count", counterResponse.Count).
		Msg("Данные о количестве успешно получены")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{
		"count": counterResponse.Count,
	})
}

func (handler *CounterMessageHandler) ResetCountByDialog(w http.ResponseWriter, r *http.Request) {
	requestID := logger.GetRequestID(r.Context())
	log := handler.getLogger(r.Context())

	senderUserID, ok := r.Context().Value("user_id").(string)
	if !ok || senderUserID == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("ID отправителя не найден в контексте")
		http.Error(w, "Текущий пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	userId := vars["user_id"]

	if userId == "" {
		log.Error().
			Str("request_id", requestID).
			Msg("Id пользователя не найден")
		http.Error(w, "userId не найден", http.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(userId)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Msg("Id пользователя невалидный")
		http.Error(w, "Id пользователя невалидный", http.StatusBadRequest)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userId).
		Str("sender_id", senderUserID).
		Msg("Запрос на сброс счетчика в микросервисе Счетчики")

	ctx := metadata.AppendToOutgoingContext(
		r.Context(),
		"x-request-id", requestID, // для логирования
		"x-user-id", senderUserID,
	)

	counterResetResponse, err := handler.service.Reset(ctx, &counter.ResetRequest{
		RecipientId: userId,
	})

	if err != nil {
		log.Error().Err(err).Msg("Не удалось сбросить счетчик в микросервисе Счетчики через gRPC")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Не удалось сбросить счетчик в микросервисе Счетчики через gRPC",
		})
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("recipient_id", counterResetResponse.RecipientId).
		Str("sender_id", userId).
		Msg("Счетчик успешно сброшен")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"result": "Счетчик успешно сброшен",
	})
}

func (s *CounterMessageHandler) getLogger(ctx context.Context) *logger.Logger {
	requestID := logger.GetRequestID(ctx)
	return s.log.WithRequestID(requestID)
}
