package handlers

import (
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"github.com/h0tw4t3r/stocksnipe_api/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const ExpiryTime = time.Hour * 24

var validate *validator.Validate
var userCollection *mongo.Collection

func init() {
	dbClient := db.MongoInit()
	userCollection = dbClient.Database("StockSnipeAPI").Collection("users")
}

func Register(w http.ResponseWriter, r *http.Request) {
	log.Debug("Registration attempt")
	inRequest := &registerRequest{}

	err := json.NewDecoder(r.Body).Decode(inRequest)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if errs := validate.Struct(inRequest); errs != nil {
		log.Error(errs)
		http.Error(w, errs.Error(), http.StatusUnprocessableEntity)
		return
	}

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

	res, err := userCollection.InsertOne(context.TODO(), inRequest)

	if err != nil {
		log.Error(err)
	}

	log.WithField("ID", res.InsertedID).Debug("New document added")
	w.WriteHeader(http.StatusOK)
}

// Login request structure
func Login(w http.ResponseWriter, r *http.Request) {
	log.Debug("Login attempt")
	inRequest := &loginRequest{}

	err := json.NewDecoder(r.Body).Decode(inRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user exists
	lookedUpUser := &loginRequest{}
	if err := userCollection.
		FindOne(context.TODO(), bson.D{{"login", inRequest.Login}}).
		Decode(&lookedUpUser); err != nil {
			log.Error(err)
			http.Error(w, "Wrong login or password provided", http.StatusUnauthorized)
			return
	}

	// JWT creation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ExpiryTime).Unix(),
	})

	// Pulling the secret key
	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		log.Error("Secret key not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, _ := token.SignedString([]byte(secretKey))

	resBytes, err := json.Marshal(&loginResponse{inRequest.Login, tokenString})
	if err != nil {
		log.WithField("err", err).Error("JSON Marshal() error")
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resBytes); err != nil {
		log.Error(err)
	}
}
