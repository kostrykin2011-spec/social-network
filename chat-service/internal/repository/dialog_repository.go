package repository

import (
	"chat-service/internal/helpers"
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type DialogRepository interface {
	CreateDialog(ctx context.Context, user1Id, user2Id uuid.UUID) (uuid.UUID, error)
	GetIdByUserId(ctx context.Context, user1Id, user2Id uuid.UUID) (uuid.UUID, error)
}

type dialogRepository struct {
	DB *sql.DB
}

func InitDialogRepository(DB *sql.DB) DialogRepository {
	return &dialogRepository{
		DB: DB,
	}
}

func (repository *dialogRepository) CreateDialog(ctx context.Context, user1Id, user2Id uuid.UUID) (uuid.UUID, error) {
	query := `INSERT INTO dialogs 
		(id, user1_id, user2_id)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	dialogId := uuid.New()
	var createdAt time.Time
	err := repository.DB.QueryRowContext(ctx, query,
		dialogId,
		user1Id,
		user2Id,
	).Scan(&createdAt)

	if err != nil {
		return uuid.Nil, err
	}

	return dialogId, nil
}

func (repository *dialogRepository) GetIdByUserId(ctx context.Context, user1Id, user2Id uuid.UUID) (uuid.UUID, error) {
	user1Id, user2Id = helpers.OrderUUIDs(user1Id, user2Id)
	query := `
    SELECT COALESCE(
        (SELECT id FROM dialogs WHERE user1_id = $1 AND user2_id = $2), 
        '00000000-0000-0000-0000-000000000000'::uuid
    )
`
	var dialogId uuid.UUID

	err := repository.DB.QueryRowContext(ctx, query, user1Id, user2Id).Scan(&dialogId)
	if err != nil {
		return uuid.Nil, err
	}

	if dialogId == uuid.Nil {
		return uuid.Nil, nil // Диалог не найден
	}

	return dialogId, nil
}
