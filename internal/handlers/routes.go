package handlers

import (
	"social-network/internal/config"
	"social-network/pkg/service"

	"github.com/gorilla/mux"
)

type Routes struct {
	config            *config.Config
	ProfileHandler    ProfileHandler
	AuthHandler       AuthHandler
	FriendShipHandler FriendsHandler
	PostHandler       PostHandler
	GenerateHandler   GenerateHandler
}

func InitRoutes(config *config.Config, authService service.AuthService, profileService service.ProfileService, friendfiendShipService service.FriendShipService, postService service.PostService) *Routes {
	return &Routes{
		config:            config,
		ProfileHandler:    InitUserHandler(profileService),
		AuthHandler:       InitAuthHandler(authService),
		FriendShipHandler: InitFriendShipHandler(friendfiendShipService),
		PostHandler:       InitPostHandler(postService),
		GenerateHandler:   InitGenerateHandler(authService, friendfiendShipService, postService),
	}
}

func (route *Routes) Run() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/login", route.AuthHandler.Login).Methods("POST")
	router.HandleFunc("/user/register", route.AuthHandler.UserRegister).Methods("POST")
	router.HandleFunc("/user/get/{id}", AuthMiddleware(route.config, route.ProfileHandler.GetProfile)).Methods("GET")
	router.HandleFunc("/user/search", AuthMiddleware(route.config, route.ProfileHandler.SearchProfile)).Methods("GET")
	router.HandleFunc("/friend/add/{id}", AuthMiddleware(route.config, route.FriendShipHandler.AddFriend)).Methods("POST")
	router.HandleFunc("/friend/delete/{id}", AuthMiddleware(route.config, route.FriendShipHandler.DeleteFriend)).Methods("DELETE")
	router.HandleFunc("/post/create", AuthMiddleware(route.config, route.PostHandler.AddPost)).Methods("POST")
	router.HandleFunc("/post/get/{id}", AuthMiddleware(route.config, route.PostHandler.GetPost)).Methods("GET")
	router.HandleFunc("/post/delete/{id}", AuthMiddleware(route.config, route.PostHandler.DeletePost)).Methods("PUT")
	router.HandleFunc("/post/feed", AuthMiddleware(route.config, route.PostHandler.GetFeed)).Methods("GET")
	router.HandleFunc("/post/feed/count", AuthMiddleware(route.config, route.PostHandler.GetFeedCount)).Methods("GET")
	router.HandleFunc("/generate/data", route.GenerateHandler.GenerateData).Methods("GET")

	return router
}
