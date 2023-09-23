package model

type Users struct {
	UserID                 string                   `json:"userId"`
	Username               string                   `json:"name"`
	ProfileImage           string                   `json:"profileImage"`
	Company                string                   `json:"company"`
	Email                  string                   `json:"email"`
	JobTitle               string                   `json:"jobTitle"`
	FacebookLink           string                   `json:"facebookLink"`
	InstagramLink          string                   `json:"instagramLink"`
	LinkedInLink           string                   `json:"linkedinLink"`
	CalendlyLink           string                   `json:"calendlyLink"`
	Mentor                 bool                     `json:"mentor"`
	WelcomeText            string                   `json:"welcomeText"`
	ProfessionalExperience string                   `json:"professionalExperience"`
	Language               []string                 `json:"language"`
	Skill                  []string                 `json:"skill"`
	Experience             float32                  `json:"experience"`
	AreaOfExperience       string                   `json:"areaOfExperience"`
	CountryDescription     []CountryDescriptionData `json:"countryDescription"`
	MentorsTopics          []MentorsTopicsData      `json:"mentorsTopics"`
	Price                  string                   `json:"price"`
	IndustryExpertise      []string                 `json:"industryExpertise"`
}

type CountryDescriptionData struct {
	Country     string `json:"country"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type MentorsTopicsData struct {
	Topic       string `json:"topic"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}
