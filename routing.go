package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/h0tw4t3r/stocksnipe_api/handlers"
	"github.com/rs/cors"
)

func Router() http.Handler  {
	r := mux.NewRouter()
	r.HandleFunc("/users/authenticate", handlers.Login)

	return cors.New(cors.Options{
		AllowedOrigins: []string{ "http://localhost:3000" },
	}).Handler(r)
}
