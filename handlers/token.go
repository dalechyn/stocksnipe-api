package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/h0tw4t3r/stocksnipe_api/tokens"
	log "github.com/sirupsen/logrus"
)

func Token(w http.ResponseWriter, r *http.Request) {
	log.Debug("RefreshToken request attempt")

	// Decoding incoming token renewal request
	inRequest := &tokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(inRequest); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := tokens.GetUserID(inRequest.RefreshToken)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If found, generating new token pair
	newAccessToken, newRefreshToken, err := tokens.UpdateTokens(userID)
	if err != nil {
		log.Error(err)
		http.Error(w, "RefreshToken pair update failed", http.StatusInternalServerError)
	}

	// JSON encoding response with new token pair
	resBytes, err := json.Marshal(&tokenResponse{newAccessToken, newRefreshToken})
	if err != nil {
		log.Error(err)
	}

	// http Response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resBytes); err != nil {
		log.Error(err)
	}
}

