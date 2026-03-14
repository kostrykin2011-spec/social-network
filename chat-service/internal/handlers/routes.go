package handlers

import (
	"chat-service/internal/services"

	"github.com/gorilla/mux"
)

type Routes struct {
	MessageHandler RedisMessageHandler
}

func InitRoutes(messageService *services.RedisMessageService) *Routes {
	return &Routes{
		MessageHandler: InitRedisMessageHanler(messageService),
	}
}

func (route *Routes) Run() *mux.Router {
	router := mux.NewRouter()
	//router.HandleFunc("/dialog/{user_id}/send", route.MessageHandler.SendMessage).Methods("POST")
	//router.HandleFunc("/dialog/{user_id}/list", route.MessageHandler.GetMessagesByUser).Methods("GET")

	return router
}
