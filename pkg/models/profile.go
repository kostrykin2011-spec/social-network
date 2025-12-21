package models

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Профиль пользователя
type Profile struct {
	Id        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Birthdate time.Time `json:"birth_date"`
	Gender    string    `json:"gender"`
	Biography string    `json:"biography"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"created_at"`
}

type ProfileResponse struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
	Biography string `json:"biography"`
	City      string `json:"city"`
}

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Birthdate string `json:"birthdate"`
	Gender    string `json:"gender"`
	Biography string `json:"biography"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type RegisterResponse struct {
	UserId string `json:"user_id"`
}

func SendSuccessResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(message)
}
