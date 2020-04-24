package handlers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// Cabinet catches GET requests only
func Cabinet(w http.ResponseWriter, r *http.Request) {
	log.Debug("Cabinet attempt")

	// Checking access token expiry
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	// if token is not provided
	if len(auth) != 2 && auth[0] != "Bearer" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// checking if token is valid and not expired

}
