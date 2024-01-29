package model

type Payload struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type ChatgptHttpPayload struct {
	Request string `json:"request"`
}

type MentorForRequest struct {
	MentorId           string               `json:"id"`
	AreaOfExpertise    []AreaOfExpertise    `json:"areaOfExpertise,omitemptyw"`
	Company            string               `json:"company,omitempty"`
	CountryDescription []CountryDescription `json:"countryDescription,omitempty"`
	Prices             []Price              `json:"prices,omitempty"`
	IndustryExpertise  []string             `json:"industryExpertise,omitempty"`
	JobTitle           string               `json:"jobTitle,omitempty"`
	Language           []string             `json:"language,omitempty"`
	MentorsTopics      []MentorsTopics      `json:"mentorsTopics,omitempty"`
	Name               string               `json:"name,omitempty"`
	Skill              []string             `json:"skill,omitempty"`
	WelcomeText        string               `json:"welcomeText,omitempty"`
	UserImage          *UserImage           `json:"userImage,omitempty"`
}
