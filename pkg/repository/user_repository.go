package repository

import (
	"context"
	"fmt"
	"social-network/pkg/database"
	"social-network/pkg/models"
	"time"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User, password string) error
	GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error)
}

type userRepository struct {
	routerDB *database.ReplicationRouter
}

func InitUserRepository(routerDB *database.ReplicationRouter) UserRepository {
	return &userRepository{routerDB: routerDB}
}

func (repository *userRepository) Create(ctx context.Context, user *models.User, password string) error {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (id, password) VALUES ($1, $2) RETURNING created_at`

	var createdAt string
	err = db.QueryRowContext(ctx, query, user.Id, password).Scan(&createdAt)

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

func (repository *userRepository) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}
	query := `select id, password from users where id = $1`

	var user models.User

	err = db.QueryRowContext(ctx, query, userId).Scan(
		&user.Id,
		&user.Password,
	)

	if err != nil {
		return nil, fmt.Errorf("Анкета не найдена")
	}

	return &user, nil
}
