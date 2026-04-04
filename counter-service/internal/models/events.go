package models

type MessageSentEvent struct {
	SagaId      string `json:"saga_id"`
	UserId      string `json:"user_id"`
	RecipientId string `json:"recipient_id"`
}
