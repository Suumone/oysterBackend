package model

import "time"

type Token struct {
	Token string `json:"token" bson:"token"`
}

type BlacklistedToken struct {
	Token     string    `bson:"token"`
	ExpiresAt time.Time `bson:"expiresAt,omitempty"`
}
