package model

type Payload struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type ChatgptHttpPayload struct {
	Request string `json:"request"`
}

type MentorForRequest struct {
	MentorId           string `json:"mentorId"`
	AreaOfExperience   string `json:"areaOfExperience"`
	Company            string `json:"company"`
	CountryDescription []struct {
		Country     string `json:"country"`
		Description string `json:"description"`
	} `json:"countryDescription"`
	Experience        int      `json:"experience"`
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
}
