package models

type WSPostFeedEvent struct {
	EventType string `json:"event_type"`
	PostID    string `json:"post_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
}
