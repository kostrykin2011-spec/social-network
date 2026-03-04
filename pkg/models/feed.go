package models

import (
	"time"

	"github.com/google/uuid"
)

type FeedItem struct {
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type PostFeedEvent struct {
	EventType string    `json:"event_type"`
	PostID    uuid.UUID `json:"post_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Text      string    `json:"text"`
	Timestamp int64     `json:"timestamp"`
}
