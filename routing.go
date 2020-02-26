package main

import (
	"net/http"

	. "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/h0tw4t3r/stocksnipe_api/handlers"
	log "github.com/sirupsen/logrus"
)

func Router() http.Handler  {
	r := mux.NewRouter()
	r.HandleFunc("/users/login", handlers.Login)
	r.HandleFunc("/users/register", handlers.Register)

	loggedHandler := LoggingHandler(log.New().Writer(), r)

	return CORS(
		AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		AllowedOrigins([]string{"http://localhost:3000"}),
		AllowedMethods([]string{"GET", "HEAD", "POST", "OPTIONS"}))(loggedHandler)
}
