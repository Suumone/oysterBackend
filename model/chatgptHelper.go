package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Payload struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type ChatgptHttpPayload struct {
	Request string `json:"request"`
}

type MentorForRequest struct {
	MentorId        string `json:"id"`
	AreaOfExpertise []struct {
		Area       string `json:"area"`
		Experience int    `json:"experience"`
	} `json:"areaOfExpertise"`
	Company            string `json:"company"`
	CountryDescription []struct {
		Country     string `json:"country"`
		Description string `json:"description"`
	} `json:"countryDescription"`
	Prices []struct {
		Price string `json:"price"`
	} `json:"prices"`
	IndustryExpertise []string `json:"industryExpertise"`
	JobTitle          string   `json:"jobTitle"`
	Language          []string `json:"language"`
	MentorsTopics     []struct {
		Description string `json:"description"`
		Topic       string `json:"topic"`
	} `json:"mentorsTopics"`
	Name        string   `json:"name"`
	Skill       []string `json:"skill"`
	WelcomeText string   `json:"welcomeText"`
	UserImage   struct {
		UserId    primitive.ObjectID `json:"userId" bson:"userId"`
		Image     string             `json:"image" bson:"image"`
		Extension string             `json:"extension" bson:"extension"`
	} `json:"userImage"`
}
