package router

import "net/http"

type (
	// Route hold configuration of routing
	Route struct {
		Desc    string
		Path    string
		Method  string
		Handler http.HandlerFunc
	}
)
