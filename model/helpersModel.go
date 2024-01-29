package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Token struct {
	Token string `json:"token" bson:"token"`
}

type BlacklistedToken struct {
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expiresAt,omitempty"`
}

type PasswordChange struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type Auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	AsMentor bool   `json:"asMentor,omitempty"`
}

type RequestParams struct {
	Fields []string `json:"fields"`
}

type UserBestMentors struct {
	BestMentors []primitive.ObjectID `bson:"bestMentors"`
}

type ValuesToSelect struct {
	Name   string   `json:"name" bson:"name"`
	Values []string `json:"values" bson:"values"`
}
