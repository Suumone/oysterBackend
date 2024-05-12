package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"oysterProject/utils"
	"time"
)

type Review struct {
	ReviewId     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MenteeId     primitive.ObjectID `json:"menteeId,omitempty" bson:"menteeId,omitempty"`
	MentorId     primitive.ObjectID `json:"mentorId,omitempty" bson:"mentorId,omitempty"`
	Review       string             `json:"review" bson:"review"`
	Rating       int                `json:"rating" bson:"rating"`
	Date         *time.Time         `json:"date" bson:"date"`
	ForFrontPage bool               `json:"forFrontPage" bson:"forFrontPage"`
	IsPublic     bool               `json:"isPublic" bson:"isPublic"`
	SessionId    primitive.ObjectID `json:"sessionId,omitempty" bson:"sessionId,omitempty"`
}

func (review *Review) FillDefaultsSessionReview(session *Session) {
	review.Date = utils.TimePtr(time.Now())
	review.ForFrontPage = false
	review.IsPublic = false
	review.SessionId = session.SessionId
	review.MentorId = session.MentorId
	review.MenteeId = session.MenteeId
}

func (review *Review) FillDefaultsMentorReview() {
	review.Date = utils.TimePtr(time.Now())
	review.ForFrontPage = true
	review.IsPublic = true
}

type UserWithReviews struct {
	MentorId primitive.ObjectID `json:"mentorId" bson:"_id"`
	Reviews  []Reviews          `json:"reviews"`
}

type Reviews struct {
	Reviewer struct {
		MenteeId  primitive.ObjectID `json:"menteeId"`
		Name      string             `json:"name"`
		JobTitle  string             `json:"jobTitle"`
		UserImage *UserImage         `json:"userImage,omitempty"`
	} `json:"reviewer"`
	Review string    `json:"review"`
	Rating int       `json:"rating"`
	Date   time.Time `json:"date"`
}

type ReviewsForFrontPage struct {
	MentorId    primitive.ObjectID `json:"mentorId" bson:"mentorId"`
	MenteeName  string             `json:"menteeName"`
	JobTitle    string             `json:"jobTitle,omitempty"`
	MenteeImage *UserImage         `json:"menteeImage,omitempty"`
	Review      string             `json:"review"`
	Rating      int                `json:"rating,omitempty"`
	Date        time.Time          `json:"date,omitempty"`
	MenteeId    primitive.ObjectID `json:"menteeId" bson:"menteeId"`
}
