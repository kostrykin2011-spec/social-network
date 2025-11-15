package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/models"
	"social-network/pkg/service"
	"social-network/pkg/utils"
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
	var registerError = models.ErrorResponse{}
	if r.Method != http.MethodPost {
		registerError.Error = "Метод не найден"
		models.SendErrorResponse(w, registerError, http.StatusMethodNotAllowed)
		return
	}

	var request models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		registerError.Error = "Невалидные данные"
		models.SendErrorResponse(w, registerError, http.StatusBadRequest)
		return
	}

	profile, err := authHandler.authService.UserRegister(&request)
	if err != nil {
		registerError.Error = err.Error()
		models.SendErrorResponse(w, registerError, http.StatusNotFound)
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
	var loginError = models.ErrorResponse{}
	if r.Method != http.MethodPost {
		loginError.Error = "Метод не найден"
		models.SendErrorResponse(w, loginError, http.StatusMethodNotAllowed)
		return
	}

	var request models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		loginError.Error = "Невалидные данные"
		models.SendErrorResponse(w, loginError, http.StatusBadRequest)
		return
	}

	user, err := authHandler.authService.Login(request.Id, request.Password)
	if err != nil {
		loginError.Error = "Пользователь не найден"
		models.SendErrorResponse(w, loginError, http.StatusNotFound)
		return
	}

	token := utils.GenerateToken()
	utils.SaveToken(token, user.Id.String())

	user.Password = ""

	response := models.LoginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
