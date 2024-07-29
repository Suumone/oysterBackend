package model

import (
	"time"
)

type UserAnalytics struct {
	UserId          string     `json:"userId,omitempty" bson:"userId,omitempty"`
	TimeStamp       *time.Time `json:"timeStamp" bson:"timeStamp"`
	LastVisitedPage string     `json:"lastVisitedPage" bson:"lastVisitedPage"`
}
