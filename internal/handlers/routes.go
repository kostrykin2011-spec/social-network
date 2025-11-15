package handlers

import (
	"social-network/pkg/service"

	"github.com/gorilla/mux"
)

type Routes struct {
	ProfileHandler ProfileHandler
	AuthHandler    AuthHandler
}

func InitRoutes(authService service.AuthService, profileService service.ProfileService) *Routes {
	return &Routes{
		ProfileHandler: InitUserHandler(profileService),
		AuthHandler:    InitAuthHandler(authService),
	}
}

func (route *Routes) CreateRotes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/login", route.AuthHandler.Login).Methods("POST")
	router.HandleFunc("/user/register", route.AuthHandler.UserRegister).Methods("POST")
	router.HandleFunc("/user/get/{id}", AuthMiddleware(route.ProfileHandler.GetProfile)).Methods("GET")

	return router
}
