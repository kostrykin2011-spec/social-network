package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type FriendShipRepository interface {
	Add(userId uuid.UUID, friendId uuid.UUID) error
	Delete(userId uuid.UUID, friendId uuid.UUID) error
	GetFriendsByUserId(userId uuid.UUID) ([]uuid.UUID, error)
}

type friendShipRepository struct {
	DB *sql.DB
}

func InitFriendShipRepository(db *sql.DB) FriendShipRepository {
	return &friendShipRepository{DB: db}
}

// Добавление друга friendId
func (repository *friendShipRepository) Add(userId uuid.UUID, friendId uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1)`
	repository.DB.QueryRow(query, userId, friendId).Scan(&exists)
	if exists {
		return fmt.Errorf("Выбранный пользователь уже есть в списке друзей")
	}

	id := uuid.New()
	status := "pending"
	query = `INSERT INTO friendships (id, user_id, friend_id, status) VALUES ($1, $2, $3, $4) RETURNING created_at`
	_, err := repository.DB.Exec(query, id, userId, friendId, status)

	return err
}

// Удаление друга friendId
func (repository *friendShipRepository) Delete(userId uuid.UUID, friendId uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1)`
	repository.DB.QueryRow(query, userId, friendId).Scan(&exists)
	if !exists {
		return fmt.Errorf("В вашем списке друзей нет выбранного пользователя")
	}

	query = `DELETE FROM friendships WHERE (user_id = $1 AND friend_id = $2 OR user_id = $2 AND friend_id = $1)`
	_, err := repository.DB.Exec(query, userId, friendId)

	return err
}

// Возвращает id-друзей
func (repository *friendShipRepository) GetFriendsByUserId(userId uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT user_id, friend_id FROM friendships WHERE friend_id = $1 OR user_id = $1`
	rows, err := repository.DB.Query(query, userId)

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
