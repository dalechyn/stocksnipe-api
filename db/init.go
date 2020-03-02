package db

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"time"
)

const ConnectTimeout = 10 * time.Second

func MongoInit() *mongo.Client {
	log.WithField("ConnectTimeout", ConnectTimeout).Info("Attempting to connect to MongoDB")
	ctx, _ := context.WithTimeout(context.Background(), ConnectTimeout)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to MongoDB")
	return client
}