package api

import (
	"github.com/volam1999/gomail/internal/app/api/handler"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/email", handler.GetAllEmail).Methods("GET")
	router.HandleFunc("/email/{id}", handler.GetEmailById).Methods("GET")
	router.HandleFunc("/email/send", handler.SendEmail).Methods("POST")
	return router
}
