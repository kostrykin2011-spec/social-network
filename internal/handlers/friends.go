package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type FriendsHandler interface {
	AddFriend(w http.ResponseWriter, r *http.Request)
	DeleteFriend(w http.ResponseWriter, r *http.Request)
}

type fiendShipHandler struct {
	service service.FriendShipService
}

func InitFriendShipHandler(service service.FriendShipService) FriendsHandler {
	return &fiendShipHandler{service: service}
}

func (handler *fiendShipHandler) AddFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	currentUserId, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Идентификатор авторизованного пользователя указан некорректно", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	friendID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Идентификатор друга указан некорректно", http.StatusBadRequest)
		return
	}

	err = handler.service.AddFiend(currentUserId, friendID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Выбранный пользователь успешно добавлен в список друзей"})
}

func (handler *fiendShipHandler) DeleteFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	currentUserId, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Идентификатор авторизованного пользователя указан некорректно", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	friendID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Идентификатор друга указан некорректно", http.StatusBadRequest)
		return
	}

	err = handler.service.Delete(currentUserId, friendID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Выбранный пользователь успешно удален из списка друзей"})
}
