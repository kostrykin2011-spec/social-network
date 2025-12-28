package main

import (
	"database/sql"
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

	masterDB, err := database.InitDatabase(config.GetConnectString(config.DatabaseConfig.Master), database.MasterDb)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД Мастер: %v", err)
	}
	defer masterDB.Close()

	var replicas []*sql.DB

	replica1DB, err := database.InitDatabase(config.GetConnectString(config.DatabaseConfig.Replica1), database.ReplicaDb)
	if err != nil {
		log.Fatalf("%v", config.GetConnectString(config.DatabaseConfig.Replica1))
		log.Fatalf("Не удалось подключиться к реплике №1: %v", err)
	} else {
		replicas = append(replicas, replica1DB)
		defer replica1DB.Close()
	}

	replica2DB, err := database.InitDatabase(config.GetConnectString(config.DatabaseConfig.Replica2), database.ReplicaDb)
	if err != nil {
		log.Fatalf("Не удалось подключиться к реплике №2: %v", err)
	} else {
		replicas = append(replicas, replica2DB)
		defer replica2DB.Close()
	}

	// Роутер баз данных
	routerDB := database.InitReplicationRouter(masterDB, replicas...)

	err = cache.InitRedis(config)

	if err != nil {
		panic(err)
	}

	defer cache.Close()

	userRepository := repository.InitUserRepository(routerDB)
	profileRepository := repository.InitProfileRepository(routerDB)
	friendShipRepository := repository.InitFriendShipRepository(routerDB)
	postRepository := repository.InitPostRepository(routerDB)

	// Инициализация кеша ленты
	feedCache := feed.NewFeedCache()
	feedService := service.InitFeedService(feedCache, postRepository, friendShipRepository)

	authService := service.InitAuthService(config, userRepository, profileRepository)
	userService := service.InitProfileService(profileRepository)
	friendShipService := service.InitFriendShipService(userRepository, friendShipRepository, feedService)
	postService := service.InitPostService(postRepository, userRepository, friendShipRepository, feedService)

	routes := handlers.InitRoutes(config, authService, userService, friendShipService, postService, routerDB)
	router := routes.Run()

	server := &http.Server{
		Addr:    ":" + config.ServerConfig.Port,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
