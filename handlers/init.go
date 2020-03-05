package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/h0tw4t3r/stocksnipe_api/db"

	"go.mongodb.org/mongo-driver/mongo"
)

var validate *validator.Validate
var userCollection *mongo.Collection

func init() {
	// Initializing dbClient for further package usage
	dbClient := db.MongoInit().Database("StockSnipeAPI")
	userCollection = dbClient.Collection("users")
}
