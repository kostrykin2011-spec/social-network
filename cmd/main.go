package main

import (
	"log"
	"net/http"
	"social-network/internal/config"
	"social-network/internal/handlers"
	"social-network/pkg/database"
	"social-network/pkg/repository"
	"social-network/pkg/service"
)

func main() {
	config := config.InitConfig()
	db, err := database.InitDatabase(config.GetConnectString())

	if err != nil {
		panic(err)
	}

	defer db.Close()

	userRepository := repository.InitUserRepository(db)
	profileRepository := repository.InitProfileRepository(db)

	authService := service.InitAuthService(userRepository, profileRepository)
	userService := service.InitProfileService(profileRepository)

	routes := handlers.InitRoutes(authService, userService)
	router := routes.CreateRotes()

	server := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
