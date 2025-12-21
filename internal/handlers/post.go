package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/models"
	"social-network/pkg/service"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PostHandler interface {
	AddPost(w http.ResponseWriter, r *http.Request)
	GetPost(w http.ResponseWriter, r *http.Request)
	DeletePost(w http.ResponseWriter, r *http.Request)
	GetFeed(w http.ResponseWriter, r *http.Request)
	GetFeedCount(w http.ResponseWriter, r *http.Request)
}

type postHandler struct {
	postService service.PostService
}

func InitPostHandler(postService service.PostService) PostHandler {
	return &postHandler{postService: postService}
}

func (handler *postHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	var postRequest models.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&postRequest); err != nil {
		models.SendErrorResponse(w, "Невалидные данные", http.StatusBadRequest)
		return
	}

	currentUserId, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Идентификатор авторизованного пользователя указан некорректно", http.StatusForbidden)
		return
	}

	err = handler.postService.AddPost(currentUserId, &postRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пост успешно создан"})
}

func (handler *postHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	postId, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Id-поста не найден", http.StatusBadRequest)
		return
	}

	var post *models.Post
	post, err = handler.postService.GetById(postId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.PostResponse{
		UserId:   post.UserId.String(),
		Title:    post.Title,
		Content:  post.Content,
		IsPublic: post.IsPublic,
	})
}

func (handler *postHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Метод не найден", http.StatusMethodNotAllowed)
		return
	}

	currentUserId, err := uuid.Parse(r.Header.Get("X-User-ID"))
	if err != nil {
		http.Error(w, "Идентификатор авторизованного пользователя указан некорректно", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	postId, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Id-поста не найден", http.StatusBadRequest)
		return
	}

	err = handler.postService.DeletePost(postId, currentUserId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пост успешно удален"})
}

func (handler *postHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentUserId, err := uuid.Parse(r.Header.Get("X-User-ID"))

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Не задан лимит", http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Не задан offset", http.StatusBadRequest)
		return
	}

	posts, err := handler.postService.GetFeed(r.Context(), currentUserId, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (handler *postHandler) GetFeedCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	currentUserId, _ := uuid.Parse(r.Header.Get("X-User-ID"))

	result := handler.postService.GetFeedCount(r.Context(), currentUserId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
