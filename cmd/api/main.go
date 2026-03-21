package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"social-network/internal/cache"
	chat "social-network/internal/client"
	"social-network/internal/config"
	"social-network/internal/feed"
	"social-network/internal/handlers"
	"social-network/internal/queue"
	"social-network/internal/worker"
	"social-network/pkg/database"
	"social-network/pkg/logger"
	"social-network/pkg/repository"
	"social-network/pkg/service"
	"syscall"
	"time"
)

func main() {
	log := logger.New(logger.Config{
		ServiceName: "monolith",
		Environment: "development",
		PrettyPrint: true,
	})

	log.Info().Msg("Запуск монолитного приложения")
	config := config.InitConfig()
	log.Info().Str("port", config.ServerConfig.Port).Msg("Конфигурация загружена")

	log.Info().Msg("Подключение к мастер БД...")
	masterDB, err := database.InitDatabase(config.GetConnectDBString(config.DatabaseConfig.Master), database.MasterDb)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось подключиться к БД Мастер")
	}
	defer masterDB.Close()

	log.Info().Msg("Подключено к мастер БД")

	log.Info().Msg("Подключение к репликам")
	replicaDB, err := database.InitDatabase(config.GetConnectDBString(config.DatabaseConfig.Replica), database.ReplicaDb)
	if err != nil {
		log.Error().Err(err).Msg("Не удалось подключиться к репликам")
	} else {
		log.Info().Msg("Подключено к репликам")
		defer replicaDB.Close()
	}

	// Роутер баз данных
	routerDB := database.InitDBRouter(masterDB, replicaDB)

	log.Info().Msg("Подключение к RabbitMQ...")
	rabbitMQ, err := queue.InitRabbitMQClient(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось инициализировать RabbitMQ")
	}
	defer rabbitMQ.Close()

	if err := rabbitMQ.CreateTaskQueue("feed_tasks"); err != nil {
		log.Fatal().Err(err).Msg("Не удалось создать очередь для задач")
	}
	log.Info().Msg("Очередь 'feed_tasks' создана")

	log.Info().Msg("Подключение к Redis...")
	err = cache.InitRedis(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось подключиться к Redis")
	}
	defer cache.Close()
	log.Info().Msg("Подключено к Redis")

	log.Info().Msg("Подключение к микросервису чатов (gRPC)...")
	grpcChatClient := chat.InitChatClient()
	if grpcChatClient == nil {
		log.Warn().Msg("Микросервис чатов недоступен, функции чата будут ограничены")
	} else {
		log.Info().Msg("Подключено к микросервису чатов")
	}

	log.Info().Msg("Инициализация репозиториев...")
	userRepository := repository.InitUserRepository(routerDB)
	profileRepository := repository.InitProfileRepository(routerDB)
	friendShipRepository := repository.InitFriendShipRepository(routerDB)
	postRepository := repository.InitPostRepository(routerDB)
	log.Info().Msg("Репозитории инициализированы")

	log.Info().Msg("Инициализация сервисов...")
	feedCache := feed.NewFeedCache()
	feedService := service.InitFeedService(feedCache, postRepository, friendShipRepository, rabbitMQ)

	authService := service.InitAuthService(config, userRepository, profileRepository)
	userService := service.InitProfileService(profileRepository)
	friendShipService := service.InitFriendShipService(userRepository, friendShipRepository, feedService)
	postService := service.InitPostService(postRepository, userRepository, friendShipRepository, feedService, rabbitMQ)
	log.Info().Msg("Сервисы инициализированы")

	log.Info().Msg("Запуск воркера обработки задач...")
	feedWorker := worker.NewFeedWorker(postRepository, feedService, rabbitMQ, "feed_tasks")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := feedWorker.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Ошибка в воркере")
		}
	}()
	log.Info().Msg("Воркер запущен")

	log.Info().Msg("Инициализация HTTP маршрутов...")
	routes := handlers.InitRoutes(
		config,
		authService,
		userService,
		friendShipService,
		postService,
		grpcChatClient,
		routerDB,
		log,
	)
	router := routes.Run()

	server := &http.Server{
		Addr:         ":" + config.ServerConfig.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", config.ServerConfig.Port).Msg("Сервер запущен")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Ошибка запуска сервера")
		}
	}()

	<-done
	log.Info().Msg("Получен сигнал завершения, останавливаем сервер...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Error().Err(err).Msg("Ошибка при остановке сервера")
	}

	log.Info().Msg("Сервер остановлен")
}
