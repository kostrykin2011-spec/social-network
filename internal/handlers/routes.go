package handlers

import (
	"context"
	"net/http"
	"social-network/internal/config"
	"social-network/pkg/database"
	"social-network/pkg/logger"
	"social-network/pkg/pb/chat"
	"social-network/pkg/service"
	"social-network/pkg/utils"
	"strings"

	"github.com/gorilla/mux"
)

type Routes struct {
	config            *config.Config
	ProfileHandler    ProfileHandler
	AuthHandler       AuthHandler
	FriendShipHandler FriendsHandler
	PostHandler       PostHandler
	GenerateHandler   GenerateHandler
	TestHandler       TestHandler
	ChatHandler       ChatHandler
	WebsocketHandler  *WebSocketHandler
	log               *logger.Logger
}

func InitRoutes(config *config.Config, authService service.AuthService, profileService service.ProfileService, friendfiendShipService service.FriendShipService, postService service.PostService, grpcChatClient chat.ChatServiceClient, routerDB *database.DBRouter, log *logger.Logger) *Routes {
	return &Routes{
		config:            config,
		ProfileHandler:    InitUserHandler(profileService),
		AuthHandler:       InitAuthHandler(authService),
		FriendShipHandler: InitFriendShipHandler(friendfiendShipService),
		PostHandler:       InitPostHandler(postService),
		GenerateHandler:   InitGenerateHandler(routerDB, authService, friendfiendShipService, postService),
		ChatHandler:       InitChatHandler(grpcChatClient, log),
		TestHandler:       InitTestHandler(routerDB),
		log:               log,
	}
}

func (route *Routes) Run() *mux.Router {
	router := mux.NewRouter()
	// Добавляем middleware для логирования ВСЕХ запросов
	//router.Use(route.loggingMiddleware)
	//router.Use(route.recoveryMiddleware)

	router.HandleFunc("/login", route.AuthHandler.Login).Methods("POST")
	router.HandleFunc("/user/register", route.AuthHandler.UserRegister).Methods("POST")
	router.HandleFunc("/user/get/{id}", route.ProfileHandler.GetProfile).Methods("GET")
	router.HandleFunc("/user/search", route.ProfileHandler.SearchProfile).Methods("GET")
	router.HandleFunc("/friend/add/{id}", route.AuthMiddleware(route.config, route.FriendShipHandler.AddFriend)).Methods("POST")
	router.HandleFunc("/friend/delete/{id}", route.AuthMiddleware(route.config, route.FriendShipHandler.DeleteFriend)).Methods("DELETE")
	router.HandleFunc("/post/create", route.AuthMiddleware(route.config, route.PostHandler.AddPost)).Methods("POST")
	router.HandleFunc("/post/get/{id}", route.AuthMiddleware(route.config, route.PostHandler.GetPost)).Methods("GET")
	router.HandleFunc("/post/delete/{id}", route.AuthMiddleware(route.config, route.PostHandler.DeletePost)).Methods("PUT")
	router.HandleFunc("/post/feed", route.AuthMiddleware(route.config, route.PostHandler.GetFeed)).Methods("GET")
	router.HandleFunc("/post/feed/count", route.AuthMiddleware(route.config, route.PostHandler.GetFeedCount)).Methods("GET")
	router.HandleFunc("/dialog/{user_id}/send", route.AuthMiddleware(route.config, route.ChatHandler.Send)).Methods("POST")
	router.HandleFunc("/dialog/{user_id}/list", route.AuthMiddleware(route.config, route.ChatHandler.GetMessages)).Methods("GET")
	router.HandleFunc("/generate/users", route.GenerateHandler.GenerateUsers).Methods("GET")

	return router
}

// Middleware для авторизации
func (route *Routes) AuthMiddleware(config *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log := route.getLogger(req.Context())

		token := req.Header.Get("Authorization")
		if token == "" {
			log.FromContext(req.Context()).Warn().
				Msg("Пользователь не авторизован")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(token, "Bearer ")
		if tokenString == token {
			log.FromContext(req.Context()).Warn().
				Msg("Пользователь не авторизован")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateToken(tokenString, config)
		if err != nil {
			log.FromContext(req.Context()).Warn().
				Err(err).
				Msg("Пользователь не авторизован")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if claims.UserID.String() == "" {
			log.FromContext(req.Context()).Error().
				Msg("Пользователь не авторизован")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(req.Context(), "user_id", claims.UserID.String())
		next.ServeHTTP(w, req.WithContext(ctx))
	}
}

// Middleware для логирования запросов
func (route *Routes) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get("X-Request-ID")
		ctx := logger.SetRequestID(req.Context(), requestID)
		w.Header().Set("X-Request-ID", logger.GetRequestID(ctx))

		log := route.getLogger(ctx)
		log.Info().
			Str("method", req.Method).
			Msg("Запрос")

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, req.WithContext(ctx))

		log.Info().
			Int("status", wrapped.status).
			Msg("Ответ")
	})
}

func (route *Routes) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log := route.getLogger(req.Context())

				log.Error().
					Interface("panic", err).
					Str("path", req.URL.Path).
					Msg("Не удалось восстановить")

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

// Вспомогательная структура для захвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

func (route *Routes) getLogger(ctx context.Context) *logger.Logger {
	requestID := logger.GetRequestID(ctx)
	if requestID != "" {
		return route.log.WithRequestID(requestID)
	}
	return route.log
}
