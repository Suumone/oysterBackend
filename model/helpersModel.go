package model

import "time"

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
}