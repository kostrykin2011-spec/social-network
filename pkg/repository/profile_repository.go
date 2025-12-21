package repository

import (
	"database/sql"
	"fmt"
	"social-network/pkg/models"
	"time"

	"github.com/google/uuid"
)

type ProfileRepository interface {
	Create(profile *models.Profile) error
	GetById(id uuid.UUID) (*models.Profile, error)
	GetByUserId(userId uuid.UUID) (*models.Profile, error)
	SearchProfiles(firstName, lastName string, limit, offset int) ([]*models.Profile, error)
}

type profileRepository struct {
	DB *sql.DB
}

func InitProfileRepository(db *sql.DB) ProfileRepository {
	return &profileRepository{DB: db}
}

func (repository *profileRepository) Create(profile *models.Profile) error {
	query := `INSERT INTO profiles (id, user_id, first_name, last_name, birth_date, gender, biography, city) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`

	var createdAt string

	err := repository.DB.QueryRow(query,
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

func (repository *profileRepository) GetById(id uuid.UUID) (*models.Profile, error) {
	query := `select id, user_id, first_name, last_name, birth_date, gender, biography, city, created_at from profiles where id = $1`

	var (
		profile   models.Profile
		createdAt string
	)

	err := repository.DB.QueryRow(query, id).Scan(
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

func (repository *profileRepository) GetByUserId(userId uuid.UUID) (*models.Profile, error) {
	query := `select id, user_id, first_name, last_name, birth_date, gender, biography, city, created_at from profiles where user_id = $1`

	var (
		profile   models.Profile
		createdAt string
	)

	err := repository.DB.QueryRow(query, userId).Scan(
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

func (repository *profileRepository) SearchProfiles(firstName, lastName string, limit, offset int) ([]*models.Profile, error) {
	firstNameForQuery := firstName + "%"
	lastNameForQuery := lastName + "%"

	query := `SELECT id, user_id, last_name, first_name, birth_date, gender, biography, city, created_at FROM profiles 
		WHERE last_name LIKE $1 and first_name LIKE $2 ORDER BY id LIMIT 10;
		`

	rows, err := repository.DB.Query(query, firstNameForQuery, lastNameForQuery)
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
