package repository

import (
	"chat-service/internal/domain"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(ctx context.Context, message *domain.Message) error
	GetMessagesByUserId(ctx context.Context, dialogId, userId uuid.UUID, limit, offset int) ([]domain.MessageResponse, error)
}

type messageRepository struct {
	DB *sql.DB
}

func InitMessageRepository(DB *sql.DB) MessageRepository {
	return &messageRepository{
		DB: DB,
	}
}

func (repository *messageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `INSERT INTO messages 
		(id, dialog_id, sender_id, recipient_id, content, user1_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`
	message.Id = uuid.New()

	var createdAt string
	err := repository.DB.QueryRowContext(ctx, query,
		message.Id,
		message.DialogId,
		message.SenderId,
		message.RecipientId,
		message.Content,
		message.User1Id).Scan(&createdAt)

	return err
}

func (repository *messageRepository) GetMessagesByUserId(ctx context.Context, dialogId, user1Id uuid.UUID, limit, offset int) ([]domain.MessageResponse, error) {
	query := `SELECT sender_id, recipient_id, content
				FROM messages 
				WHERE dialog_id = $1 AND user1_id = $2
				ORDER BY created_at DESC
				LIMIT $3 OFFSET $4`

	rows, err := repository.DB.QueryContext(ctx, query, dialogId, user1Id, limit, offset)
	if err != nil {
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	var messages []domain.MessageResponse
	for rows.Next() {
		var message domain.MessageResponse
		err := rows.Scan(&message.SenderId, &message.RecipientId, &message.Content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
