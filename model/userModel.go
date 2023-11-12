package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id                     primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Username               string               `json:"name" bson:"name,omitempty"`
	ProfileImageId         primitive.ObjectID   `json:"-" bson:"profileImageId,omitempty"`
	Company                string               `json:"company" bson:"company,omitempty"`
	Email                  string               `json:"email" bson:"email,omitempty"`
	JobTitle               string               `json:"jobTitle" bson:"jobTitle,omitempty"`
	FacebookLink           string               `json:"facebookLink" bson:"facebookLink,omitempty"`
	InstagramLink          string               `json:"instagramLink" bson:"instagramLink,omitempty"`
	LinkedInLink           string               `json:"linkedinLink" bson:"linkedinLink,omitempty"`
	CalendlyLink           string               `json:"calendlyLink" bson:"calendlyLink,omitempty"`
	IsMentor               bool                 `json:"isMentor" bson:"isMentor"`
	WelcomeText            string               `json:"welcomeText" bson:"welcomeText,omitempty"`
	ProfessionalExperience string               `json:"professionalExperience" bson:"professionalExperience,omitempty"`
	Language               []string             `json:"language" bson:"language,omitempty"`
	Skill                  []string             `json:"skill" bson:"skill,omitempty"`
	Experience             float32              `json:"experience" bson:"experience,omitempty"`
	AreaOfExpertise        string               `json:"areaOfExpertise" bson:"areaOfExpertise,omitempty"`
	CountryDescription     []CountryDescription `json:"countryDescription" bson:"countryDescription,omitempty"`
	MentorsTopics          []MentorsTopics      `json:"mentorsTopics" bson:"mentorsTopics,omitempty"`
	Prices                 []Price              `json:"prices" bson:"prices,omitempty"`
	IndustryExpertise      []string             `json:"industryExpertise" bson:"industryExpertise,omitempty"`
	Password               string               `json:"-" bson:"password,omitempty"`
	IsNewUser              bool                 `json:"isNewUser" bson:"isNewUser"`
	IsApproved             bool                 `json:"isApproved" bson:"isApproved,omitempty"`
	IsTopMentor            bool                 `json:"isTopMentor" bson:"isTopMentor,omitempty"`
	AsMentor               bool                 `json:"asMentor" bson:"asMentor,omitempty"`
	UserImage              UserImageResult      `json:"userImage" bson:"userImage,omitempty"`
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
