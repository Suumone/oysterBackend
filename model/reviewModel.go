package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserWithReviews struct {
	MentorId primitive.ObjectID `json:"mentorId" bson:"_id"`
	Reviews  []Reviews          `json:"reviews"`
}

type Reviews struct {
	Reviewer struct {
		MenteeId  primitive.ObjectID `json:"menteeId"`
		Name      string             `json:"name"`
		JobTitle  string             `json:"jobTitle"`
		UserImage UserImageResult    `json:"userImage,omitempty"`
	} `json:"reviewer"`
	Review string    `json:"review"`
	Rating int       `json:"rating"`
	Date   time.Time `json:"date"`
}

type ReviewsForFrontPage struct {
	MentorId    primitive.ObjectID `json:"mentorId" bson:"mentorId"`
	MenteeName  string             `json:"menteeName"`
	JobTitle    string             `json:"jobTitle,omitempty"`
	MenteeImage UserImageResult    `json:"menteeImage,omitempty"`
	Review      string             `json:"review"`
	Rating      int                `json:"rating,omitempty"`
	Date        time.Time          `json:"date,omitempty"`
	MenteeId    primitive.ObjectID `json:"menteeId" bson:"menteeId"`
}
