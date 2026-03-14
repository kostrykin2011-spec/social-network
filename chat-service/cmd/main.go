package main

import (
	"chat-service/config"
	"chat-service/internal/cache"
	"chat-service/internal/helpers"
	"chat-service/internal/logger"
	"chat-service/internal/services"
	"context"
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"os/signal"
	"social-network/pkg/pb/chat"
	"syscall"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Создаем логгер
	log := logger.New(logger.Config{
		ServiceName: "chat-service",
		Environment: "development",
		PrettyPrint: true,
	})

	log.Info().Msg("Запуск микросервиса диалогов")

	config := config.InitConfig()
	log.Info().Str("port", config.ServerConfig.Port).Msg("Конфигурация загружена")

	rdb := cache.InitRedis()
	defer rdb.Close()
	log.Info().Msg("Redis подключен")

	redisService := services.InitRedisMessageService(rdb)
	chatService := services.NewChatService(redisService, log)

	lis, err := net.Listen("tcp", ":"+config.ServerConfig.Port)
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка запуска TCP")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(log)),
	)

	chat.RegisterChatServiceServer(grpcServer, chatService)
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

func loadTestData() {
	filePath := "/app/users.csv"
	if filePath == "" {
		fmt.Println("Файл users.csv не найден")
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении CSV:", err)
	}

	globalUUIDCache := helpers.Instance()
	for _, record := range records {
		userId, _ := uuid.Parse(record[0])
		globalUUIDCache.Add(userId)
	}
}
