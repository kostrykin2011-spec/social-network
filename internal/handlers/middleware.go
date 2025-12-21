package handlers

import (
	"net/http"
	"social-network/internal/config"
	"social-network/pkg/models"
	"social-network/pkg/utils"
	"strings"
)

func AuthMiddleware(config *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			models.SendErrorResponse(w, "Токен не найден", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utils.ValidateToken(tokenString, config)
		if err != nil {
			http.Error(w, "Токен не валидный", http.StatusUnauthorized)

			return
		}

		// Добавляем user_id в контекст запроса
		r.Header.Set("X-User-ID", claims.UserID.String())

		next(w, r)
	}
}
