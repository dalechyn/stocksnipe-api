package models

import "go.mongodb.org/mongo-driver/x/mongo/driver/uuid"

type TokenSchema struct {
	TokenId uuid.UUID `bson:"tokenID, omitempty"`
	UserId string `bson:"userID, omitempty"`
}
