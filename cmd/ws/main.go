package main

import (
	"log"
	"net/http"
	"social-network/internal/config"
	"social-network/internal/handlers"
	"social-network/internal/hub"
	"social-network/internal/queue"
	"time"

	"github.com/google/uuid"
)

func main() {
	config := config.InitConfig()
	serverId := "ws-" + uuid.New().String()[:8]
	var rabbitMQ *queue.RabbitMQClient
	var err error

	// Пробуем подключиться (5 попыток)
	for i := 0; i < 5; i++ {
		rabbitMQ, err = queue.InitRabbitMQClient(config)
		if err == nil {
			break
		}

		time.Sleep(3 * time.Second)
	}

	if rabbitMQ == nil {
		log.Fatal("Не удалось подключиться к очереди RabbitMQ")
	}
	defer rabbitMQ.Close()

	log.Printf("RabbitMQ - успешно подключен")
	wsHub := hub.InitWebsocketHub(rabbitMQ, serverId)
	go wsHub.Run()
	defer wsHub.Stop()

	wsHandler := handlers.NewWebSocketHandler(wsHub)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/feed", wsHandler.HandleWebSocket)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Для корневого пути отдаем index.html
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./frontend/index.html")
			return
		}

		// Пробуем найти файл в папке frontend
		http.ServeFile(w, r, "./frontend/"+r.URL.Path)
	})

	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Не удалось запустить веб-сокет сервер №"+serverId, err)
	}
}
