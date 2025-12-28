package repository

import (
	"context"
	"fmt"
	"social-network/pkg/database"

	"github.com/google/uuid"
)

type FriendShipRepository interface {
	Add(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error
	Delete(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error
	GetFriendsByUserId(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error)
}

type friendShipRepository struct {
	routerDB *database.ReplicationRouter
}

func InitFriendShipRepository(routerDB *database.ReplicationRouter) FriendShipRepository {
	return &friendShipRepository{routerDB: routerDB}
}

// Добавление друга friendId
func (repository *friendShipRepository) Add(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return err
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1)`
	db.QueryRowContext(ctx, query, userId, friendId).Scan(&exists)
	if exists {
		return fmt.Errorf("Выбранный пользователь уже есть в списке друзей")
	}

	id := uuid.New()
	status := "pending"
	query = `INSERT INTO friendships (id, user_id, friend_id, status) VALUES ($1, $2, $3, $4) RETURNING created_at`
	_, err = db.ExecContext(ctx, query, id, userId, friendId, status)

	return err
}

// Удаление друга friendId
func (repository *friendShipRepository) Delete(ctx context.Context, userId uuid.UUID, friendId uuid.UUID) error {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return err
	}

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1)`
	db.QueryRowContext(ctx, query, userId, friendId).Scan(&exists)
	if !exists {
		return fmt.Errorf("В вашем списке друзей нет выбранного пользователя")
	}

	query = `DELETE FROM friendships WHERE (user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1)`
	_, err = db.ExecContext(ctx, query, userId, friendId)

	return err
}

// Возвращает id-друзей
func (repository *friendShipRepository) GetFriendsByUserId(ctx context.Context, userId uuid.UUID) ([]uuid.UUID, error) {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	query := `SELECT user_id, friend_id FROM friendships WHERE friend_id = $1 OR user_id = $1`
	rows, err := db.QueryContext(ctx, query, userId)

	if err != nil {
		return nil, err
	}

	if rows == nil {
		return nil, nil
	}

	var friendIds []uuid.UUID
	for rows.Next() {
		var (
			val1 uuid.UUID
			val2 uuid.UUID
		)
		err := rows.Scan(&val1, &val2)
		if err != nil {
			return nil, err
		}
		if val1 != userId {
			friendIds = append(friendIds, val1)
		} else {
			friendIds = append(friendIds, val2)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return friendIds, nil
}
