package repository

import (
	"database/sql"
	"fmt"
	"social-network/pkg/models"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *models.User, password string) error
	GetUserById(userId uuid.UUID) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func InitUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (repository *userRepository) Create(user *models.User, password string) error {
	query := `INSERT INTO users (id, password) VALUES ($1, $2) RETURNING created_at`

	var createdAt string
	err := repository.db.QueryRow(query,
		user.Id,
		password).Scan(&createdAt)

	if err != nil {
		return err
	}

	parsedTime, err := time.Parse("2006-01-02", createdAt)
	if err != nil {
		parsedTime = time.Now()
	}
	user.CreatedAt = parsedTime

	return nil
}

func (repository *userRepository) GetUserById(userId uuid.UUID) (*models.User, error) {
	query := `select id, password from users where id = $1`

	var user models.User

	err := repository.db.QueryRow(query, userId).Scan(
		&user.Id,
		&user.Password,
	)

	if err != nil {
		return nil, fmt.Errorf("Анкета не найдена")
	}

	return &user, nil
}
