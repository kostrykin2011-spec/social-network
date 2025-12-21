package models

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Пользователь
type User struct {
	Id        uuid.UUID `json:"id"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthRequest struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token  string    `json:"token"`
	UserId uuid.UUID `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func SendErrorResponse(w http.ResponseWriter, error string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(error)
}
