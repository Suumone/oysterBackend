package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id                     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Username               string               `json:"name,omitempty" bson:"name,omitempty"`
	ProfileImageId         primitive.ObjectID   `json:"-" bson:"profileImageId,omitempty"`
	Company                string               `json:"company" bson:"company,omitempty"`
	Email                  string               `json:"email" bson:"email,omitempty"`
	JobTitle               string               `json:"jobTitle" bson:"jobTitle,omitempty"`
	FacebookLink           string               `json:"facebookLink" bson:"facebookLink,omitempty"`
	InstagramLink          string               `json:"instagramLink" bson:"instagramLink,omitempty"`
	LinkedInLink           string               `json:"linkedinLink" bson:"linkedinLink,omitempty"`
	WelcomeText            string               `json:"welcomeText" bson:"welcomeText,omitempty"`
	ProfessionalExperience string               `json:"professionalExperience" bson:"professionalExperience,omitempty"`
	Language               []string             `json:"language,omitempty" bson:"language,omitempty"`
	Skill                  []string             `json:"skill,omitempty" bson:"skill,omitempty"`
	Experience             int32                `json:"-" bson:"experience,omitempty"`
	AreaOfExpertise        []AreaOfExpertise    `json:"areaOfExpertise,omitempty" bson:"areaOfExpertise,omitempty"`
	CountryDescription     []CountryDescription `json:"countryDescription,omitempty" bson:"countryDescription,omitempty"`
	MentorsTopics          []MentorsTopics      `json:"mentorsTopics,omitempty" bson:"mentorsTopics,omitempty"`
	Prices                 []Price              `json:"prices,omitempty" bson:"prices,omitempty"`
	IndustryExpertise      []string             `json:"industryExpertise,omitempty" bson:"industryExpertise,omitempty"`
	Password               string               `json:"-" bson:"password,omitempty"`
	IsNewUser              bool                 `json:"isNewUser" bson:"isNewUser"`
	IsApproved             bool                 `json:"isApproved" bson:"isApproved,omitempty"`
	IsTopMentor            bool                 `json:"isTopMentor" bson:"isTopMentor,omitempty"`
	AsMentor               bool                 `json:"asMentor" bson:"asMentor,omitempty"`
	UserImage              *UserImage           `json:"userImage,omitempty" bson:"userImage,omitempty"`
	UserMentorRequest      string               `json:"userMentorRequest" bson:"userMentorRequest,omitempty"`
	Availability           []*Availability      `json:"availability,omitempty" bson:"availability,omitempty"`
	MeetingLink            string               `json:"meetingLink" bson:"meetingLink,omitempty"`
	UserRegisterDate       *time.Time           `json:"userRegisterDate" bson:"userRegisterDate,omitempty"`
	LatestTimeZone         int                  `json:"latestTimeZone" bson:"latestTimeZone"`
}

type CountryDescription struct {
	Country     string `json:"country" bson:"country,omitempty"`
	Description string `json:"description" bson:"description,omitempty"`
}

type MentorsTopics struct {
	Topic       string `json:"topic" bson:"topic,omitempty"`
	Description string `json:"description" bson:"description,omitempty"`
}

type Price struct {
	Price string `json:"price" bson:"price,omitempty"`
}

type AreaOfExpertise struct {
	Area       string `json:"area" bson:"area,omitempty"`
	Experience int32  `json:"experience" bson:"experience,omitempty"`
}

type UserState struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AsMentor bool               `json:"asMentor" bson:"asMentor,omitempty"`
}

type UserImage struct {
	UserId         primitive.ObjectID `json:"userId" bson:"userId"`
	Name           string             `json:"name,omitempty" bson:"name,omitempty"`
	Email          string             `json:"-" bson:"email"`
	Image          []byte             `json:"image" bson:"image"`
	Extension      string             `json:"extension" bson:"extension"`
	LatestTimeZone int                `json:"-" bson:"latestTimeZone"`
}

type Availability struct {
	Weekday  string `json:"weekday" bson:"weekday"`
	TimeFrom string `json:"timeFrom" bson:"timeFrom"`
	TimeTo   string `json:"timeTo" bson:"timeTo"`
	TimeZone int32  `json:"timeZone" bson:"timeZone"`
}
