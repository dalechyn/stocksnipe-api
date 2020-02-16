package StockSnipe_API

import (
	"github.com/gorilla/mux"
	"github.com/h0tw4t3r/stocksnipe_api"
)

func Router() {
	r := mux.NewRouter()
	r.HandleFunc("/users/authenticate", handlers.Login())
}
