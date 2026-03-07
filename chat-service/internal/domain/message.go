package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id          uuid.UUID `json:"id"`
	DialogId    uuid.UUID `json:"dialog_id"`
	SenderId    uuid.UUID `json:"sender_id"`
	RecipientId uuid.UUID `json:"recipient_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User1Id     uuid.UUID `json:"user1_id"`
}

type MessageRequest struct {
	Content string `json:"content"`
}

type MessageResponse struct {
	SenderId    uuid.UUID `json:"sender_id"`
	RecipientId uuid.UUID `json:"recipient_id"`
	Content     string    `json:"content"`
}

type RedisMessageResponse struct {
	Content string `json:"content"`
}
