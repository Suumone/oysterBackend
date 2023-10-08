package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserWithReviews struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Reviews []struct {
		Reviewer struct {
			Id           primitive.ObjectID `json:"id"`
			Name         string             `json:"name"`
			JobTitle     string             `json:"jobTitle"`
			ProfileImage string             `json:"profileImage,omitempty"`
		} `json:"reviewer"`
		Review string    `json:"review"`
		Rating int       `json:"rating"`
		Date   time.Time `json:"date"`
	} `json:"reviews"`
}

type ReviewsForFrontPage struct {
	Name         string    `json:"name"`
	JobTitle     string    `json:"jobTitle,omitempty"`
	ProfileImage string    `json:"profileImage,omitempty"`
	Review       string    `json:"review"`
	Rating       int       `json:"rating,omitempty"`
	Date         time.Time `json:"date,omitempty"`
}
