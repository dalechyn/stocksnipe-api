package models

type UserSchema struct {
	ID      string `bson:"_id,omitempty"`
	Email    string `bson:"email,omitempty"`
	Login    string `bson:"login,omitempty"`
	Password string `bson:"password,omitempty"`
}