package models

import (
	"time"

	"github.com/google/uuid"
)

// Посты
type Post struct {
	Id        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdatePostRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	IsPublic bool   `json:"is_public"`
}

type PostResponse struct {
	UserId   string `json:"user_id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	IsPublic bool   `json:"is_public"`
}
