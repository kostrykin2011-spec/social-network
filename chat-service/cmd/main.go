package main

import (
	"chat-service/config"
	"chat-service/internal/handlers"
	"chat-service/internal/helpers"
	"chat-service/internal/repository"
	"chat-service/internal/services"
	"chat-service/pkg/database"
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

	db, err := database.InitDatabase(config.GetConnectString(config.DBConfig))
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer db.Close()

	messageRepository := repository.InitMessageRepository(db)
	dialogRepository := repository.InitDialogRepository(db)

	filePath := "/app/users.csv"
	if filePath == "" {
		fmt.Println("Файл people.csv не найден")
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
	
	messageService := services.InitMessageService(dialogRepository, messageRepository)

	routes := handlers.InitRoutes(messageService)
	router := routes.Run()

	server := &http.Server{
		Addr:    ":" + config.ServerConfig.Port,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

}
