package repository

import (
	"database/sql"
	"fmt"
	"social-network/pkg/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PostRepository interface {
	AddPost(post *models.Post) error
	GetById(postId uuid.UUID) (*models.Post, error)
	DeletePost(postId, userId uuid.UUID) error
	GetListByUserId(userId uuid.UUID, limit, offset int) ([]*models.Post, error)
	GetListByUserIds(userIds []uuid.UUID, limit, offset int) ([]*models.Post, error)
}

type postRepository struct {
	DB *sql.DB
}

func InitPostRepository(db *sql.DB) PostRepository {
	return &postRepository{DB: db}
}

func (repository *postRepository) AddPost(post *models.Post) error {
	var createdAt string
	query := `INSERT INTO posts (id, user_id, title, content, is_public) VALUES ($1, $2, $3, $4, $5) RETURNING created_at`

	err := repository.DB.QueryRow(query,
		&post.Id,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.IsPublic).Scan(&createdAt)

	createdAtValue, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return err
	}
	post.CreatedAt = createdAtValue

	return nil
}

func (repository *postRepository) DeletePost(postId, userId uuid.UUID) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1 AND user_id = $2)`
	err := repository.DB.QueryRow(query, postId, userId).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("Пост не найден")
	}

	query = `DELETE FROM posts WHERE id = $1`
	_, err = repository.DB.Exec(query, postId)

	return err
}

func (repository *postRepository) GetById(postId uuid.UUID) (*models.Post, error) {
	var post models.Post
	query := `SELECT id, user_id, title, content, is_public, created_at FROM posts WHERE id = $1`
	err := repository.DB.QueryRow(query, postId).Scan(
		&post.Id,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.IsPublic,
		&post.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("Пост не найден")
	}

	return &post, err
}

// Получить все посты определенного пользователя
func (repository *postRepository) GetListByUserId(userId uuid.UUID, limit, offset int) ([]*models.Post, error) {
	query := `SELECT id, user_id, title, content, is_public, created_at FROM posts WHERE user_id = $1 ORDER BY created_at DESC limit $2 OFFSET $3`
	rows, err := repository.DB.Query(query, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.Id, &post.UserId, &post.Title, &post.Content, &post.IsPublic, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (repository *postRepository) GetListByUserIds(userIds []uuid.UUID, limit, offset int) ([]*models.Post, error) {
	placeholders := make([]string, len(userIds))
	args := make([]interface{}, len(userIds))

	for i, u := range userIds {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = u.String()
	}

	query := fmt.Sprintf(
		"SELECT id, user_id, title, content, is_public, created_at FROM posts WHERE user_id::uuid IN (%s) ORDER BY created_at DESC OFFSET %s LIMIT %s",
		strings.Join(placeholders, ", "),
		strconv.Itoa(offset),
		strconv.Itoa(limit),
	)

	rows, err := repository.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.Id, &post.UserId, &post.Title, &post.Content, &post.IsPublic, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (repository *postRepository) buildInQuery(uuids []uuid.UUID) (string, []interface{}) {
	// Создаем плейсхолдеры: $1, $2, $3...
	placeholders := make([]string, len(uuids))
	args := make([]interface{}, len(uuids))

	for i, u := range uuids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = u.String() // Преобразуем UUID в строку
	}

	query := fmt.Sprintf(
		"SELECT id, user_id, title, content, is_public, created_at FROM posts WHERE user_id::uuid IN (%s)",
		strings.Join(placeholders, ", "),
	)

	return query, args
}
