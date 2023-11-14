package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Session struct {
	MentorId          primitive.ObjectID `json:"mentorId" bson:"mentorId"`
	MenteeId          primitive.ObjectID `json:"menteeId" bson:"menteeId"`
	SessionTime       primitive.DateTime `json:"sessionTime" bson:"sessionTime"`
	LengthMinutes     int                `json:"lengthMinutes" bson:"lengthMinutes"`
	RequestFromMentee string             `json:"requestFromMentee" bson:"requestFromMentee"`
	IsConfirmed       bool               `json:"isConfirmed" bson:"isConfirmed"`
}
