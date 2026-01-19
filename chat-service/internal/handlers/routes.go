package handlers

import (
	"chat-service/internal/services"

	"github.com/gorilla/mux"
)

type Routes struct {
	MessageHandler MessageHandler
}

func InitRoutes(messageService services.MessageService) *Routes {
	return &Routes{
		MessageHandler: InitMessageHanler(messageService),
	}
}

func (route *Routes) Run() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/dialog/{user_id}/send", route.MessageHandler.Send).Methods("POST")
	router.HandleFunc("/dialog/{user_id}/list", route.MessageHandler.GetMessages).Methods("GET")

	return router
}
