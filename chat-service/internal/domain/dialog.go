package domain

import (
	"time"

	"github.com/google/uuid"
)

type Dialog struct {
	Id        uuid.UUID `json:"id"`
	User1Id   uuid.UUID `json:"user1_id"`
	User2Id   uuid.UUID `json:"user2_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdateAt  time.Time `json:"update_at"`
}
