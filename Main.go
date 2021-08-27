package main

import (
	"main/internal/app/api"
	"main/internal/pkg/config/envconfig"
	"main/internal/pkg/log"
	"net/http"
)

func main() {
	if envconfig.Error != nil {
		log.Fatal("Error loading .env file")
		return
	}
	// go checkScheduleEmail()
	router := api.NewRouter()
	http.Handle("/", router)
	http.ListenAndServe(":8080", nil)
}
