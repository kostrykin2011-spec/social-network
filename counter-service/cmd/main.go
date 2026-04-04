package main

import (
	"context"
	"counter-service/cache"
	"counter-service/config"
	"counter-service/internal/consumer"
	"counter-service/internal/repository"
	"counter-service/internal/services"
	logger "counter-service/pkg"
	"counter-service/pkg/database"
	"net"
	"os"
	"os/signal"
	"social-network/pkg/pb/counter"
	"syscall"

	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func main() {
	log := logger.New(logger.Config{
		ServiceName: "counter-service",
	})
	log.Info().Msg("Запуск микросервиса счетчиков")

	config := config.InitConfig()
	log.Info().Str("port", config.ServerConfig.Port).Msg("Конфигурация загружена")

	log.Info().Msg("Подключение к Redis...")
	redisClient, err := cache.InitRedis(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось подключиться к Redis")
	}
	log.Info().Msg("Подключено к Redis")
	defer cache.Close()

	log.Info().Msg("Подключение к RabbitMQ...")
	rabbitMQConnect, err := amqp091.Dial(config.GetConnectRabbitMQString("rabbitmq"))
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось подключиться к RabbitMQ")
	}
	log.Info().Msg("Подключено к RabbitMQ")
	defer rabbitMQConnect.Close()

	db, err := database.InitDatabase(config.GetDBConnectString(config.DBConfig))
	if err != nil {
		log.Fatal().Err(err).Msg("Не удалось подключиться к БД")
	}
	log.Info().Msg("Подключено к БД")
	defer db.Close()

	dbRepository := repository.InitCounterDBRepository(db)
	redisRepository := repository.InitCacheCounterRepository(redisClient)

	counterService := services.InitCounterService(dbRepository, redisRepository, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumerMessageSent := consumer.InitMessageConsumer(rabbitMQConnect, redisRepository, dbRepository, log)
	go consumerMessageSent.Start(ctx)

	lis, err := net.Listen("tcp", ":"+config.ServerConfig.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка запуска TCP")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(log)),
	)

	counterGrpcService := services.NewCounterGrpcService(counterService, log)
	counter.RegisterCounterServiceServer(grpcServer, counterGrpcService)
	reflection.Register(grpcServer)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", config.ServerConfig.Port).Msg("gRPC сервер запущен")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Ошибка работы gRPC сервера")
		}
	}()

	<-done
	log.Info().Msg("Получен сигнал завершения")

	grpcServer.GracefulStop()
	log.Info().Msg("Сервер остановлен")

}

func loggingInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var requestID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-request-id"); len(ids) > 0 {
				requestID = ids[0]
			}
		}
		ctx = logger.SetRequestID(ctx, requestID)
		log.WithRequestID(requestID).Info().
			Str("method", info.FullMethod).
			Msg("gRPC запрос")

		resp, err := handler(ctx, req)

		// Логируем ответ
		if err != nil {
			log.WithRequestID(requestID).Error().
				Err(err).
				Str("method", info.FullMethod).
				Msg("gRPC ошибка")
		} else {
			log.WithRequestID(requestID).Info().
				Str("method", info.FullMethod).
				Msg("gRPC ответ")
		}

		return resp, err
	}
}
