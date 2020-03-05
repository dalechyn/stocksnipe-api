package handlers

import (
	"context"
	"encoding/json"
	"github.com/h0tw4t3r/stocksnipe_api/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	log.Debug("Registration attempt")

	// Decoding incoming request
	inRequest := &registerRequest{}
	err := json.NewDecoder(r.Body).Decode(inRequest)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validating it
	if errs := validate.Struct(inRequest); errs != nil {
		log.Error(errs)
		http.Error(w, errs.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check if any user with same email or login exists, if so, send an error
	if err := userCollection.FindOne(
		context.TODO(),
		bson.M{"$or":
			bson.A{
				bson.M{"login": inRequest.Login},
				bson.M{"email": inRequest.Email}}}).Err(); err == nil {
		log.Error(err)
		http.Error(w, "User with that login or email already exists", http.StatusConflict)
		return
	}

	// Add new user if no matches found
 	res, err := userCollection.InsertOne(context.TODO(), models.
	UserSchema{Email: inRequest.Email, Login: inRequest.Login, Password: inRequest.Password})

	if err != nil {
		log.Error(err)
	}

	log.WithField("ID", res.InsertedID).Debug("New document added")

 	// Send ok status
	w.WriteHeader(http.StatusOK)
}
