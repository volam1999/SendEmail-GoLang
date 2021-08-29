package main

import (
	"net/http"

	"github.com/volam1999/gomail/internal/app/api"
	"github.com/volam1999/gomail/internal/app/api/handler"
	"github.com/volam1999/gomail/internal/pkg/config/envconfig"
	"github.com/volam1999/gomail/internal/pkg/log"
)

func main() {
	if envconfig.Error != nil {
		log.Fatal("Error loading .env file")
		return
	}
	go handler.SendScheduleEmail()
	router := api.NewRouter()
	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
