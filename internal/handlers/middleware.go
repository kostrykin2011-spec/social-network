package handlers

import (
	"net/http"
	"social-network/pkg/models"
	"social-network/pkg/utils"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err = models.ErrorResponse{}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err.Error = "Токен не найден"
			models.SendErrorResponse(w, err, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			err.Error = "Некорректный формат токена"
			models.SendErrorResponse(w, err, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userID, isValid := utils.ValidateToken(token)
		if !isValid {
			err.Error = "Срок токена истек"
			models.SendErrorResponse(w, err, http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-User-ID", userID)

		next(w, r)
	}
}
