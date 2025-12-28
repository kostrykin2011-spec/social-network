package repository

import (
	"context"
	"database/sql"
	"fmt"
	"social-network/pkg/database"
	"social-network/pkg/models"
	"time"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *models.Profile) error
	GetById(ctx context.Context, id uuid.UUID) (*models.Profile, error)
	GetByUserId(ctx context.Context, userId uuid.UUID) (*models.Profile, error)
	SearchProfiles(ctx context.Context, firstName, lastName string, limit, offset int) ([]*models.Profile, error)
}

type profileRepository struct {
	routerDB *database.ReplicationRouter
}

func InitProfileRepository(routerDB *database.ReplicationRouter) ProfileRepository {
	return &profileRepository{routerDB: routerDB}
}

func (repository *profileRepository) Create(ctx context.Context, profile *models.Profile) error {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return err
	}

	query := `INSERT INTO profiles (id, user_id, first_name, last_name, birth_date, gender, biography, city) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`

	var createdAt string

	err = db.QueryRowContext(ctx, query,
		profile.Id,
		profile.UserId,
		profile.FirstName,
		profile.LastName,
		profile.Birthdate,
		profile.Gender,
		profile.Biography,
		profile.City,
	).Scan(&createdAt)

	if err != nil {
		return fmt.Errorf("Не удалось создать анкету пользователя: %w", err)
	}

	parsedCreatedAt, _ := time.Parse("2006-01-02 15:04:05", createdAt)

	profile.CreatedAt = parsedCreatedAt

	return nil
}

func (repository *profileRepository) GetById(ctx context.Context, id uuid.UUID) (*models.Profile, error) {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	query := `select id, user_id, first_name, last_name, birth_date, gender, biography, city, created_at from profiles where id = $1`

	var (
		profile   models.Profile
		createdAt string
	)

	err = db.QueryRowContext(ctx, query, id).Scan(
		&profile.Id,
		&profile.UserId,
		&profile.FirstName,
		&profile.LastName,
		&profile.Birthdate,
		&profile.Gender,
		&profile.Biography,
		&profile.City,
		&createdAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Анкета не найдена")
	}

	parsedCreatedAt, _ := time.Parse("2006-01-02 15:04:05", createdAt)
	profile.CreatedAt = parsedCreatedAt

	return &profile, nil
}

func (repository *profileRepository) GetByUserId(ctx context.Context, userId uuid.UUID) (*models.Profile, error) {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}
	query := `select id, user_id, first_name, last_name, birth_date, gender, biography, city, created_at from profiles where user_id = $1`

	var (
		profile   models.Profile
		createdAt string
	)

	err = db.QueryRowContext(ctx, query, userId).Scan(
		&profile.Id,
		&profile.UserId,
		&profile.FirstName,
		&profile.LastName,
		&profile.Birthdate,
		&profile.Gender,
		&profile.Biography,
		&profile.City,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Анкета не найдена")
	}

	if err != nil {
		return nil, fmt.Errorf("Анкета не найдена: %w", err)
	}

	parsedCreatedAt, _ := time.Parse("2006-01-02 15:04:05", createdAt)

	profile.CreatedAt = parsedCreatedAt

	return &profile, nil
}

func (repository *profileRepository) SearchProfiles(ctx context.Context, firstName, lastName string, limit, offset int) ([]*models.Profile, error) {
	db, err := repository.routerDB.GetDatabase(ctx)
	if err != nil {
		return nil, err
	}

	firstNameForQuery := firstName + "%"
	lastNameForQuery := lastName + "%"

	query := `SELECT id, user_id, last_name, first_name, birth_date, gender, biography, city, created_at FROM profiles 
		WHERE last_name LIKE $1 and first_name LIKE $2 ORDER BY id LIMIT 10;
		`

	rows, err := db.QueryContext(ctx, query, firstNameForQuery, lastNameForQuery)
	if err != nil {
		return nil, err
	}

	var profiles []*models.Profile
	for rows.Next() {
		var (
			profile   models.Profile
			createdAt time.Time
		)
		err := rows.Scan(
			&profile.Id,
			&profile.UserId,
			&profile.FirstName,
			&profile.LastName,
			&profile.Birthdate,
			&profile.Gender,
			&profile.Biography,
			&profile.City,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, &profile)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}
