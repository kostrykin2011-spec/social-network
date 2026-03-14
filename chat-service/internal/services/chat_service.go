package services

import (
	"chat-service/internal/domain"
	"chat-service/internal/logger"
	"context"
	"social-network/pkg/pb/chat"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ChatService struct {
	chat.UnimplementedChatServiceServer
	redis *RedisMessageService
	log   *logger.Logger
}

func NewChatService(redis *RedisMessageService, log *logger.Logger) *ChatService {
	return &ChatService{
		redis: redis,
		log:   log,
	}
}

func (s *ChatService) getRequestID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ids := md.Get("x-request-id"); len(ids) > 0 {
			return ids[0]
		}
	}
	return ""
}

func (s *ChatService) getLogger(ctx context.Context) *logger.Logger {
	return s.log.WithRequestID(s.getRequestID(ctx))
}

func (s *ChatService) getMetadata(ctx context.Context) (userID string, requestID string) {
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

func (s *ChatService) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.MessageResponse, error) {
	userID, requestID := s.getMetadata(ctx)
	log := s.getLogger(ctx)

	log.Info().
		Str("user_id", userID).
		Str("friend_id", req.RecipientId).
		Msg("Запрос в микросервисе на отправку сообщения")

	senderID, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Невалидный user_id в metadata")
		return nil, status.Error(codes.InvalidArgument, "Невалидный user_id в metadata")
	}

	friendID, err := uuid.Parse(req.RecipientId)
	if err != nil {
		log.Error().Err(err).Str("friend_id", req.RecipientId).Msg("Невалидный friend_id")
		return nil, status.Error(codes.InvalidArgument, "Невалидный friend_id")
	}

	message := domain.MessageRequest{
		Content: req.Content,
	}

	if err := s.redis.Send(ctx, message, senderID, friendID); err != nil {
		log.Error().Err(err).Msg("Ошибка сохранения в Redis Cluster")
		return nil, status.Error(codes.Internal, "Ошибка сохранения сообщения в Redis Cluster")
	}

	log.Info().
		Str("x-request-id", requestID).
		Str("sender_id", senderID.String()).
		Str("friend_id", friendID.String()).
		Msg("Сообщение сохранено в Redis Cluster")

	return &chat.MessageResponse{
		SenderId:    senderID.String(),
		RecipientId: friendID.String(),
		Content:     req.Content,
	}, nil
}

func (s *ChatService) GetMessagesByUser(ctx context.Context, req *chat.GetDialogRequest) (*chat.DialogResponse, error) {
	userID, requestID := s.getMetadata(ctx)
	log := s.getLogger(ctx)

	log.Info().
		Str("x-request-id", requestID).
		Str("user_id", userID).
		Str("friend_id", req.FriendId).
		Msg("Получен запрос на получение сообщений")

	currentUserID, err := uuid.Parse(userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Невалидный user_id в метаданных gRPC")
		return nil, status.Error(codes.InvalidArgument, "Невалидный user_id в метаданных gRPC")
	}

	friendID, err := uuid.Parse(req.FriendId)
	if err != nil {
		log.Error().Err(err).Str("friend_id", req.FriendId).Msg("Невалидный friend_id")
		return nil, status.Error(codes.InvalidArgument, "Невалидный friend_id")
	}

	messages, err := s.redis.GetMessagesByUser(ctx, currentUserID, friendID, int(req.Limit), int(req.Offset))
	if err != nil {
		log.Error().Err(err).Msg("Ошибка получения сообщений из диалога - Redis Cluster")
		return nil, status.Error(codes.Internal, "Ошибка получения сообщений из диалога - Redis Cluster")
	}

	var protoMessages []*chat.Message
	for _, msg := range messages {
		protoMessages = append(protoMessages, &chat.Message{
			SenderId:    currentUserID.String(),
			RecipientId: friendID.String(),
			Content:     msg.Content,
		})
	}

	return &chat.DialogResponse{
		Messages: protoMessages,
	}, nil
}
