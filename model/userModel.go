package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id                     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Username               string               `json:"name,omitempty" bson:"name,omitempty"`
	ProfileImageId         primitive.ObjectID   `json:"-" bson:"profileImageId,omitempty"`
	Company                string               `json:"company,omitempty" bson:"company,omitempty"`
	Email                  string               `json:"email,omitempty" bson:"email,omitempty"`
	JobTitle               string               `json:"jobTitle,omitempty" bson:"jobTitle,omitempty"`
	FacebookLink           string               `json:"facebookLink,omitempty" bson:"facebookLink,omitempty"`
	InstagramLink          string               `json:"instagramLink,omitempty" bson:"instagramLink,omitempty"`
	LinkedInLink           string               `json:"linkedinLink,omitempty" bson:"linkedinLink,omitempty"`
	CalendlyLink           string               `json:"calendlyLink,omitempty" bson:"calendlyLink,omitempty"`
	WelcomeText            string               `json:"welcomeText,omitempty" bson:"welcomeText,omitempty"`
	ProfessionalExperience string               `json:"professionalExperience,omitempty" bson:"professionalExperience,omitempty"`
	Language               []string             `json:"language,omitempty" bson:"language,omitempty"`
	Skill                  []string             `json:"skill,omitempty" bson:"skill,omitempty"`
	Experience             float32              `json:"-" bson:"experience,omitempty"`
	AreaOfExpertise        []AreaOfExpertise    `json:"areaOfExpertise,omitempty" bson:"areaOfExpertise,omitempty"`
	CountryDescription     []CountryDescription `json:"countryDescription,omitempty" bson:"countryDescription,omitempty"`
	MentorsTopics          []MentorsTopics      `json:"mentorsTopics,omitempty" bson:"mentorsTopics,omitempty"`
	Prices                 []Price              `json:"prices,omitempty" bson:"prices,omitempty"`
	IndustryExpertise      []string             `json:"industryExpertise,omitempty" bson:"industryExpertise,omitempty"`
	Password               string               `json:"-" bson:"password,omitempty"`
	IsNewUser              bool                 `json:"isNewUser" bson:"isNewUser"`
	IsApproved             bool                 `json:"isApproved,omitempty" bson:"isApproved,omitempty"`
	IsTopMentor            bool                 `json:"isTopMentor,omitempty" bson:"isTopMentor,omitempty"`
	AsMentor               bool                 `json:"asMentor" bson:"asMentor,omitempty"`
	UserImage              UserImageResult      `json:"userImage,omitempty" bson:"userImage,omitempty"`
	UserMentorRequest      string               `json:"userMentorRequest,omitempty" bson:"userMentorRequest,omitempty"`
	Availability           []Availability       `json:"availability,omitempty" bson:"availability,omitempty"`
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
	Area       string  `json:"area" bson:"area,omitempty"`
	Experience float32 `json:"experience" bson:"experience,omitempty"`
}

type UserState struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AsMentor bool               `json:"asMentor" bson:"asMentor,omitempty"`
}

type UserImage struct {
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`
	Image     [][]byte           `json:"image" bson:"image"`
	Extension string             `json:"extension" bson:"extension"`
}
type UserImageResult struct {
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`
	Image     string             `json:"image" bson:"image"`
	Extension string             `json:"extension" bson:"extension"`
}

type Availability struct {
	Weekday  string `json:"weekday" bson:"weekday"`
	TimeFrom string `json:"timeFrom" bson:"timeFrom"`
	TimeTo   string `json:"timeTo" bson:"timeTo"`
}
