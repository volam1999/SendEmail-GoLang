package api

import (
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()

	emailSrv, _ := newEmailService()
	emailHandler := newEmailHandler(emailSrv)

	for _, route := range emailHandler.Routes() {
		router.HandleFunc(route.Path, route.Handler).Methods(route.Method)
	}
	go emailHandler.SendScheduleEmail()
	// router.HandleFunc("/email", handler.GetAllEmail).Methods("GET")
	// router.HandleFunc("/email/{id}", handler.GetEmailById).Methods("GET")
	// router.HandleFunc("/email/send", handler.SendEmail).Methods("POST")
	return router
}
