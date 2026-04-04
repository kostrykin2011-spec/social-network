package chat

import (
	"log"
	"os"
	"social-network/pkg/pb/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitChatClient() chat.ChatServiceClient {
	addr := os.Getenv("CHAT_SERVICE_ADDR")
	if addr == "" {
		addr = "chat-service:5001"
	}

	// Устанавливаем соединение
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Не шифруем
	)
	if err != nil {
		log.Fatalf("Не удалось установить соединение gRPC: %v", err)
	}

	return chat.NewChatServiceClient(conn)
}
