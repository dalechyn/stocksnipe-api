package handlers

import (
	"context"
	"encoding/json"
	"github.com/h0tw4t3r/stocksnipe_api/models"
	"github.com/h0tw4t3r/stocksnipe_api/tokens"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

// Login request structure
func Login(w http.ResponseWriter, r *http.Request) {
	log.Debug("Login attempt")

	// Decoding http request
	inRequest := &loginRequest{}
	err := json.NewDecoder(r.Body).Decode(inRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validating incoming data
	if errs := validate.Struct(inRequest); errs != nil {
		log.Error(errs)
		http.Error(w, errs.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check if user exists
	lookedUpUser := &models.UserSchema{}
	if err := userCollection.
		FindOne(context.TODO(), bson.D{{"login", inRequest.Login}}).
		Decode(&lookedUpUser); err != nil {
		log.Error(err)
		http.Error(w, "Wrong login or password provided", http.StatusUnauthorized)
		return
	}

	// Generate some tokens for him
	accessToken, refreshToken, err := tokens.UpdateTokens(lookedUpUser.ID)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// Encode to JSON and send him
	resBytes, err := json.Marshal(&loginResponse{inRequest.Login, accessToken, refreshToken})
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resBytes); err != nil {
		log.Error(err)
	}
}
