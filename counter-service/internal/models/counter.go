package models

import "github.com/google/uuid"

type Counter struct {
	UserId      uuid.UUID `json:"user_id"`
	RecipientId uuid.UUID `json:"recipient_id"`
	UnreadCount int64     `json:"unread_count"`
}

type UnreadCountRequest struct {
	UserId      string `json:"user_id"`
	RecipientId string `json:"recipient_id"`
}

type UnreadCountResponse struct {
	UserId      string `json:"user_id"`
	RecipientId string `json:"recipient_id"`
	Count       int64  `json:"count"`
}

type CountResetRequest struct {
	UserId      string `json:"user_id"`
	RecipientId string `json:"recipient_id"`
}

type CountResetResponse struct {
	UserId      string `json:"user_id"`
	RecipientId string `json:"recipient_id"`
	Success     bool   `json:"success"`
}
