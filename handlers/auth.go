package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginJSON struct {
	username string
	password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	var p LoginJSON

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("Body: %+v", p)
}
