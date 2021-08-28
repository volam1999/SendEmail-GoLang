package api

import (
	"main/internal/app/api/handler"

	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/send", handler.SendHandler)
	router.HandleFunc("/get", handler.GetHandler)
	router.HandleFunc("/get/{id}", handler.GetByIdHandler)
	return router
}
