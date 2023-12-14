package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Status int

const (
	CreatedByMentee Status = iota
	PendingByMentor
	ReschedulingByMentor
	ReschedulingByMentee
	Confirmed
	Completed
	CanceledByMentor
	CanceledByMentee
	Expired
)

type Session struct {
	SessionId           primitive.ObjectID `json:"sessionId" bson:"_id,omitempty"`
	MentorId            primitive.ObjectID `json:"mentorId" bson:"mentorId"`
	MenteeId            primitive.ObjectID `json:"menteeId" bson:"menteeId"`
	SessionTimeStart    *time.Time         `json:"sessionTimeStart" bson:"sessionTimeStart"`
	SessionTimeEnd      *time.Time         `json:"sessionTimeEnd" bson:"sessionTimeEnd"`
	NewSessionTimeStart *time.Time         `json:"newSessionTimeStart,omitempty" bson:"newSessionTimeStart,omitempty"`
	NewSessionTimeEnd   *time.Time         `json:"newSessionTimeEnd,omitempty" bson:"newSessionTimeEnd,omitempty"`
	RequestFromMentee   string             `json:"requestFromMentee" bson:"requestFromMentee"`
	SessionStatus       Status             `json:"-" bson:"sessionStatus"`
	Status              string             `json:"status" bson:"-"`
	StatusForMentee     string             `json:"statusForMentee" bson:"-"`
	StatusForMentor     string             `json:"statusForMentor" bson:"-"`
	PaymentDetails      string             `json:"paymentDetails,omitempty" bson:"paymentDetails,omitempty"`
	MeetingLink         string             `json:"meetingLink,omitempty" bson:"meetingLink,omitempty"`
}

type GroupedSessions struct {
	PendingSessions  []Session `json:"pendingSessions"`
	UpcomingSessions []Session `json:"upcomingSessions"`
	PastSessions     []Session `json:"pastSessions"`
}

type AvailableWeekday struct {
	Date    time.Time `json:"date"`
	Weekday string    `json:"weekday"`
}

type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

func (s Status) GetStatusForMentee() string {
	switch s {
	case CreatedByMentee:
		return "Session created(waiting for payment)"
	case PendingByMentor:
		return "Pending confirmation from mentor"
	case ReschedulingByMentor:
		return "Awaiting your confirmation(rescheduling request from mentor)"
	case ReschedulingByMentee:
		return "Pending confirmation from mentor"
	case CanceledByMentor:
		return "Session canceled by mentor"
	case CanceledByMentee:
		return "Session canceled"
	case Confirmed:
		return "Confirmed"
	case Completed:
		return "Completed"
	case Expired:
		return "Expired"
	default:
		return "Unknown"
	}
}

func (s Status) GetStatusForMentor() string {
	switch s {
	case CreatedByMentee:
		return "Session created"
	case PendingByMentor:
		return "Awaiting your confirmation"
	case ReschedulingByMentor:
		return "Pending confirmation from mentee"
	case ReschedulingByMentee:
		return "Awaiting your confirmation(rescheduling request from mentee)"
	case CanceledByMentor:
		return "Session canceled"
	case CanceledByMentee:
		return "Session canceled by mentee"
	case Confirmed:
		return "Confirmed"
	case Completed:
		return "Completed"
	case Expired:
		return "Expired"
	default:
		return "Unknown"
	}
}

func (s Status) String() string {
	switch s {
	case CreatedByMentee:
		return "createdByMentee"
	case PendingByMentor:
		return "pendingByMentor"
	case ReschedulingByMentor:
		return "reschedulingByMentor"
	case ReschedulingByMentee:
		return "reschedulingByMentee"
	case CanceledByMentor:
		return "canceledByMentor"
	case CanceledByMentee:
		return "canceledByMentee"
	case Confirmed:
		return "confirmed"
	case Completed:
		return "completed"
	case Expired:
		return "expired"
	default:
		return "unknown"
	}
}
