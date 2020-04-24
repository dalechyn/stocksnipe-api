package handlers

import (
	"net/http"
	log "github.com/sirupsen/logrus"
)

func Test401(w http.ResponseWriter, r *http.Request) {
	log.Debug("Test401")
	w.WriteHeader(401)
}
