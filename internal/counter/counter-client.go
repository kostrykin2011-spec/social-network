package counter

import (
	"log"
	"os"
	"social-network/pkg/pb/counter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitCounterClient() counter.CounterServiceClient {
	addr := os.Getenv("COUNTER_SERVICE_ADDR")
	if addr == "" {
		addr = "counter-service:5001"
	}

	// Устанавливаем соединение
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Не шифруем
	)
	if err != nil {
		log.Fatalf("Не удалось установить соединение gRPC: %v", err)
	}

	return counter.NewCounterServiceClient(conn)
}
