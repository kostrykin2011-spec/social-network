package main

import (
	"log"
	"net/http"
	"social-network/internal/cache"
	"social-network/internal/config"
	"social-network/internal/feed"
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

	err = cache.InitRedis(config)

	if err != nil {
		panic(err)
	}

	defer cache.Close()

	userRepository := repository.InitUserRepository(db)
	profileRepository := repository.InitProfileRepository(db)
	friendShipRepository := repository.InitFriendShipRepository(db)
	postRepository := repository.InitPostRepository(db)

	// Инициализация кеша ленты
	feedCache := feed.NewFeedCache()
	feedService := service.InitFeedService(feedCache, postRepository, friendShipRepository)

	authService := service.InitAuthService(config, userRepository, profileRepository)
	userService := service.InitProfileService(profileRepository)
	friendShipService := service.InitFriendShipService(userRepository, friendShipRepository, feedService)
	postService := service.InitPostService(postRepository, userRepository, friendShipRepository, feedService)

	routes := handlers.InitRoutes(config, authService, userService, friendShipService, postService)
	router := routes.Run()

	server := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
