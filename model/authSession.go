package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Oauth2User struct {
	Name  string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
}

type AuthSession struct {
	SessionId primitive.ObjectID `bson:"_id,omitempty"`
	UserId    primitive.ObjectID `bson:"userId"`
	Expiry    int64              `bson:"expiry"`
}

func (s AuthSession) isExpired() bool {
	return s.Expiry < time.Now().Unix()
}
