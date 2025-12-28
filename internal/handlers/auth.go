package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/models"
	"social-network/pkg/service"
)

type AuthHandler interface {
	UserRegister(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	authService service.AuthService
}

func InitAuthHandler(service service.AuthService) AuthHandler {
	return &authHandler{authService: service}
}

func (authHandler *authHandler) UserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		models.SendErrorResponse(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	var request models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		models.SendErrorResponse(w, "Невалидные данные", http.StatusBadRequest)
		return
	}

	profile, err := authHandler.authService.UserRegister(r.Context(), &request)
	if err != nil {
		models.SendErrorResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	response := models.RegisterResponse{
		UserId: profile.UserId.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (authHandler *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		models.SendErrorResponse(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	var request models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		models.SendErrorResponse(w, "Невалидные данные", http.StatusBadRequest)
		return
	}

	response, err := authHandler.authService.Login(r.Context(), request.Id, request.Password)
	if err != nil {
		models.SendErrorResponse(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
