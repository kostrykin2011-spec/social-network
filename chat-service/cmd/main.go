package main

import (
	"chat-service/config"
	"chat-service/internal/cache"
	"chat-service/internal/handlers"
	"chat-service/internal/helpers"
	"chat-service/internal/services"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	config := config.InitConfig()
	rdb := cache.InitRedis()
	defer rdb.Close()
	// Проверка соединения
	messageService := services.InitRedisMessageService(rdb)

	routes := handlers.InitRoutes(messageService)
	router := routes.Run()

	// Загружаем тестовые данные
	//loadTestData()

	server := &http.Server{
		Addr:    ":" + config.ServerConfig.Port,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
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
