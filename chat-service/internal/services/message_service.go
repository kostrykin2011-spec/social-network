package services

import (
	"chat-service/internal/domain"
	"chat-service/internal/helpers"
	"chat-service/internal/repository"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MessageService interface {
	Send(ctx context.Context, messageRequest domain.MessageRequest, senderId, recipientId uuid.UUID) error
	GetMessagesByUser(ctx context.Context, user1Id, user2Id uuid.UUID, limit, offset int) ([]domain.MessageResponse, error)
}

type messageService struct {
	dialogRepository  repository.DialogRepository
	messageRepository repository.MessageRepository
}

func InitMessageService(dialogRepository repository.DialogRepository, messageRepository repository.MessageRepository) MessageService {
	return &messageService{
		dialogRepository:  dialogRepository,
		messageRepository: messageRepository,
	}
}

func (service *messageService) Send(ctx context.Context, messageRequest domain.MessageRequest, senderId, recipientId uuid.UUID) error {
	user1Id, user2Id := helpers.OrderUUIDs(senderId, recipientId)
	dialogId, err := service.dialogRepository.GetIdByUserId(ctx, user1Id, user2Id)

	if dialogId == uuid.Nil {
		dialogId, err = service.dialogRepository.CreateDialog(ctx, user1Id, user2Id)
		if err != nil {
			return fmt.Errorf("Не удалось создать диалог между пользователями: ", err.Error())
		}
	}
	message := &domain.Message{
		Id:          uuid.New(),
		DialogId:    dialogId,
		SenderId:    user1Id,
		RecipientId: user2Id,
		Content:     messageRequest.Content,
		CreatedAt:   time.Now(),
		User1Id:     user1Id,
	}

	err = service.messageRepository.Create(ctx, message)
	if err != nil {
		return fmt.Errorf("Не удалось отправить сообщение пользователю ", err.Error())
	}

	return nil
}

func (service *messageService) GetMessagesByUser(ctx context.Context, user1Id, user2Id uuid.UUID, limit, offset int) ([]domain.MessageResponse, error) {
	dialogId, err := service.dialogRepository.GetIdByUserId(ctx, user1Id, user2Id)
	if err != nil {
		return nil, err
	}
	if dialogId == uuid.Nil {
		return nil, fmt.Errorf("Диалог отсутствует")
	}
	user1Id, user2Id = helpers.OrderUUIDs(user1Id, user2Id)
	messages, err := service.messageRepository.GetMessagesByUserId(ctx, dialogId, user1Id, limit, offset)

	return messages, err
}
