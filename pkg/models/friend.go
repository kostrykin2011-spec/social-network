package models

import "github.com/google/uuid"

// Друзья
type FriendShip struct {
	Id        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	FriendId  uuid.UUID `json:"friend_id"`
	Status    string    `json:"status"`
	CreatedAt uuid.Time `json:"created_at"`
}
