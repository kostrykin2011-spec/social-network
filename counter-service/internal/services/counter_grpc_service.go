package services

import (
	"context"
	logger "counter-service/pkg"
	"social-network/pkg/pb/counter"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CounterGrpcService struct {
	counter.UnimplementedCounterServiceServer
	counterService *CounterService
	log            *logger.Logger
}

func NewCounterGrpcService(counterService *CounterService, log *logger.Logger) *CounterGrpcService {
	return &CounterGrpcService{
		counterService: counterService,
		log:            log,
	}
}

func (s *CounterGrpcService) getRequestID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ids := md.Get("x-request-id"); len(ids) > 0 {
			return ids[0]
		}
	}
	return ""
}

func (s *CounterGrpcService) getLogger(ctx context.Context) *logger.Logger {
	return s.log.WithRequestID(s.getRequestID(ctx))
}

func (s *CounterGrpcService) getMetadata(ctx context.Context) (userID, requestID string) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if users := md.Get("x-user-id"); len(users) > 0 {
			userID = users[0]
		}
		if requests := md.Get("x-request-id"); len(requests) > 0 {
			requestID = requests[0]
		} else {
			requestID = uuid.New().String()
		}
	}
	return userID, requestID
}

// Получение количества непрочитанных сообщений через gRPC
func (s *CounterGrpcService) GetCount(ctx context.Context, req *counter.GetCountRequest) (*counter.GetCountResponse, error) {
	senderId, requestID := s.getMetadata(ctx)
	log := s.getLogger(ctx)

	log.Info().
		Str("x-request-id", requestID).
		Str("sender_id", senderId).
		Str("recipient_id", req.RecipientId).
		Msg("Запрос в микросервисе Счетчики на получение количества непрочитанных сообщений")

	_, err := uuid.Parse(senderId)
	if err != nil {
		log.Error().Err(err).Str("sender_id", senderId).Msg("Невалидный sender_id")
		return nil, status.Error(codes.InvalidArgument, "Невалидный sender_id")
	}

	_, err = uuid.Parse(req.RecipientId)
	if err != nil {
		log.Error().Err(err).Str("recipient_id", req.RecipientId).Msg("Невалидный recipient_id")
		return nil, status.Error(codes.InvalidArgument, "Невалидный recipient_id")
	}

	count, err := s.counterService.GetCount(ctx, req.RecipientId, senderId)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка при получении данных о количестве непрочитанных сообщений")
		return nil, status.Error(codes.Internal, "Ошибка при получении данных о количестве непрочитанных сообщений")
	}

	return &counter.GetCountResponse{
		RecipientId: req.RecipientId,
		Count:       count,
	}, nil
}

// Сброс счетчика через gRPC
func (s *CounterGrpcService) Reset(ctx context.Context, req *counter.ResetRequest) (*counter.ResetResponse, error) {
	senderId, requestID := s.getMetadata(ctx)
	log := s.getLogger(ctx)

	log.Info().
		Str("x-request-id", requestID).
		Str("sender_id", senderId).
		Str("recipient_id", req.RecipientId).
		Msg("Получен запрос на сброс счетчика")

	_, err := uuid.Parse(req.RecipientId)
	if err != nil {
		log.Error().Err(err).Str("recipient_id", req.RecipientId).Msg("Невалидный recipient_id")
		return nil, status.Error(codes.InvalidArgument, "Невалидный recipient_id")
	}

	_, err = uuid.Parse(senderId)
	if err != nil {
		log.Error().Err(err).Str("sender_id", senderId).Msg("Невалидный sender_id")
		return nil, status.Error(codes.InvalidArgument, "Невалидный sender_id")
	}

	err = s.counterService.Reset(ctx, req.RecipientId, senderId)
	if err != nil {
		log.Error().Err(err).Msg("Ошибка при сбросе счетчика")
		return nil, status.Error(codes.Internal, "Ошибка при сбросе счетчика")
	}

	log.Info().
		Str("x-request-id", requestID).
		Str("recipient_id", req.RecipientId).
		Str("sender_id", senderId).
		Msg("Счетчик успешно сброшен")

	return &counter.ResetResponse{
		RecipientId: req.RecipientId,
		Result:      true,
	}, nil
}
