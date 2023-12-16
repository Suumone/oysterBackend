package httpHandlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sashabaranov/go-openai"
	"log"
	"net/http"
	"os"
	"oysterProject/database"
	"oysterProject/model"
	"strings"
)

const (
	OpenApiModel = openai.GPT3Dot5Turbo1106
)

var (
	ApiKey = os.Getenv("OPENAI_API_KEY")
)

func mapUserToMentorForRequest(user model.User) model.MentorForRequest {
	mentor := model.MentorForRequest{
		MentorId:    user.Id.Hex(),
		Company:     user.Company,
		JobTitle:    user.JobTitle,
		Language:    user.Language,
		Name:        user.Username,
		Skill:       user.Skill,
		WelcomeText: user.WelcomeText,
	}

	for _, countryDesc := range user.CountryDescription {
		mentor.CountryDescription = append(mentor.CountryDescription, struct {
			Country     string `json:"country"`
			Description string `json:"description"`
		}{
			Country:     countryDesc.Country,
			Description: countryDesc.Description,
		})
	}

	for _, mentorTopic := range user.MentorsTopics {
		mentor.MentorsTopics = append(mentor.MentorsTopics, struct {
			Description string `json:"description"`
			Topic       string `json:"topic"`
		}{
			Description: mentorTopic.Description,
			Topic:       mentorTopic.Topic,
		})
	}

	for _, areaOfExpertise := range user.AreaOfExpertise {
		mentor.AreaOfExpertise = append(mentor.AreaOfExpertise, struct {
			Area       string `json:"area"`
			Experience int    `json:"experience"`
		}{
			Area:       areaOfExpertise.Area,
			Experience: int(areaOfExpertise.Experience),
		})
	}

	for _, price := range user.Prices {
		mentor.Prices = append(mentor.Prices, struct {
			Price string `json:"price"`
		}{
			Price: price.Price,
		})
	}

	return mentor
}

func CalculateBestMentors(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}
	var requestPayload model.ChatgptHttpPayload
	if err := ParseJSONRequest(r, &requestPayload); err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	go database.UpdateMentorRequest(requestPayload.Request, userId)

	mentorsFromChatgpt, err := sendRequestToChatgpt(requestPayload.Request, userId)
	if err != nil {
		return
	}

	WriteJSONResponse(w, http.StatusOK, mentorsFromChatgpt)
}

func sendRequestToChatgpt(request string, userId string) ([]model.MentorForRequest, error) {
	mentors, err := database.GetMentors(nil, "")
	if err != nil {
		return nil, errors.New("error getting mentors from database")
	}
	var mentorsForRequest []model.MentorForRequest
	for _, mentor := range mentors {
		mentorsForRequest = append(mentorsForRequest, mapUserToMentorForRequest(mentor))
	}
	jsonMentorsForRequestBytes, err := json.Marshal(mentorsForRequest)
	if err != nil {
		log.Println("Error encoding to JSON:", err)
		return nil, err
	}
	jsonMentorsForRequestString := string(jsonMentorsForRequestBytes)
	systemContent := "There will be json with a list of mentors here, analyze it. Field explanation:\n \"mentorId\" - mentor id use it for response on my requests.\n\"areaOfExpertise\": The mentor's area of expertise.\n\"company\": The mentor's current employer.\n\"countryDescription\": A list of countries where the mentor has lived and worked, along with descriptions of their experiences in those countries.\n\"experience\": The mentor's total years of experience.\n\"industryExpertise\": A list of industries in which the mentor has expertise.\n\"jobTitle\": The mentor's job title.\n\"language\": The languages the mentor can communicate in.\n\"mentorsTopics\": A list of topics or areas in which the mentor can provide guidance.\n\"name\": The mentor's name.\n\"skill\": A list of skills the mentor possesses.\n\"welcomeText\": A brief introduction and welcome message from the mentor, including their background, experience and other information.\n\nList of mentors:\n" +
		jsonMentorsForRequestString
	//"[{\"id\":\"65107e2d72355f1b2610c609\",\"areaOfExpertise\":\"Operations\",\"company\":\"Yango Delivery\",\"countryDescription\":[{\"country\":\"Zambia\",\"description\":\"I am still living and working in African counties such as Zambia, Ghana, Morocco for half a year already. I can share my knowledge about obtaining local visas, local culture, work environment, renting an apartment\"},{\"country\":\"Brazil\",\"description\":\"I've been living and working in São Paulo (Brazil) for half of the year. I can share my knowledge about obtaining local visas, local culture, work environment, renting an apartment\"},{\"country\":\"United Arab Emirates\",\"description\":\"I've been living and working in Dubai (UAE) for 1.5 years. I can share my knowledge about obtaining local visas, local culture, work environment, renting an apartment\"}],\"experience\":8,\"industryExpertise\":[\"E-commerce\",\"FoodTech\",\"Ride-hailing\",\"Transportation\",\"Food Delivery\"],\"jobTitle\":\"Manager, Africa Operations\",\"language\":[\"English\",\"Russian\"],\"mentorsTopics\":[{\"description\":\"Internationalization strategy: which geography,domain and company to choose. How to apply to international companies: creation and review of CV and cover letter, preparation for the interviews, establishing and leveraging a network to get a job.\",\"topic\":\"Internationalization of your career as a manager in technological company\"},{\"description\":\"Establish a business growth strategy encompassing Business Operations Management and Efficiency, Business Developmentand Sales, Market Growth & Expansion, Business Strategies & Plans, Budgeting, Financial Modeling, Product, Customer management.\",\"topic\":\"Grow your business volumes and efficiency\"},{\"description\":\"How to sell the results of your work: build beautiful and self-sufficient slides, create a concise speech\",\"topic\":\"Sales and presentation excellence\"}],\"name\":\"Ilya Abdulkin\",\"skill\":[\"Go-to-market Strategy\",\"Product Launches\",\"Product-Market Fit\",\"Growth Strategy\",\"People Management\"],\"welcomeText\":\"Hi, my name is Ilya. I gained an extensive experience in top-tier international consulting firms and IT companies such as McKinsey&Company, Kearney, Yandex/Yango, DiDi, Mail.Ru (VK). I have managed and consulted businesses across Russia & CIS, Eastern Europe, Latin America, the Middle East, and African markets.\"},{\"id\":\"65107e2d72355f1b2610c60a\",\"areaOfExperience\":\"Corporate Development & Strategy\",\"company\":\"inDrive\",\"countryDescription\":[{\"country\":\"Cyprus\",\"description\":\"Officially employed in Cyprus and spend a lot of time working there. Know all island and have huge network of network\"},{\"country\":\"Turkey\",\"description\":\"Launched business and developedbusiness network. Know very well business landscape of Istanbul and cultural features\"},{\"country\":\"Lebanon\",\"description\":\"I have lived in Lebanon for a specific period of time, I have widenedmy horizon in the business field of the country; I have also learned about its culture, geography, law and economy, also I developed a businesss network that connects me directly to others.\"}],\"experience\":6,\"industryExpertise\":[\"Ride-hailing\"],\"jobTitle\":\"General Manager for MENA region\",\"language\":[\"Russian\",\"English\",\"Arabic\"],\"mentorsTopics\":[{\"description\":\"To know what are you good at and to be able to focus on it in career / personal life\",\"topic\":\"Identify personal strength\"},{\"description\":\"To keep up with career, physical health, mental health and relationship health with people\",\"topic\":\"How to make all around life development\"},{\"description\":\"How to choose best career track to achieve final results. What are goals in professional life and how to link them to personal goals\",\"topic\":\"Setting up career track and personal development goals\"}],\"name\":\"Max Osipov\",\"skill\":[\"Leadership\",\"Growth Strategy\",\"Financial Modeling\",\"Product Launches\",\"Go-to-marketStrategy\"],\"welcomeText\":\"Hello everyone, my name is Max Osipov and I’m your mentor for this workshop, i will be leading you all the way step by step to reachyour goals successfully, I’ll make sure you finish this course with a wide knowledge of the outs and snouts of the business field. The best of luck.\"},{\"id\":\"65107e2d72355f1b2610c60b\",\"areaOfExperience\":\"Product Management\",\"company\":\"Delivery Hero\",\"countryDescription\":[{\"country\":\"Germany\",\"description\":\"Live in berlin for ~1.5 years, leader of a russian-speaking community in Berlin (140 relocants) and essentially can talk about anything - visa, banks, apartment-hunt, life routing, etc.\"},{\"country\":\"Azerbaijan\",\"description\":\"\"}],\"experience\":8,\"jobTitle\":\"Group Product Manager\",\"language\":[\"English\",\"Russian\",\"Turkish\"],\"mentorsTopics\":[{\"description\":\"We will talk about how to figure out what to MVP is, how to make an MVP out of this MVP and then launch it on a pretty tight timeline\",\"topic\":\"Launching MVP\"},{\"description\":\"How to effectively communicate with your stakeholders\",\"topic\":\"Communication\"},{\"description\":\"How to hire, grow and motivate the entire team\",\"topic\":\"People Management\"}],\"name\":\"Shakhriiar Soltanov (Shakh Solt)\",\"skill\":[\"Building a Team\",\"Career Development\",\"CV Preparation\",\"Communication\",\"FinTech\",\"Soft Skills\",\"Stakeholders Management\",\"People Management\"],\"welcomeText\":\"For my entire adult life I always found myself sharing knowledge and loving it.\\n\\nWhen I was moving to Berlin in 2022 I found many irregularities in Visa process. It was so stupid that people were advised to follow excessive path, so I started sharing the short cuts with friends and their friend. This routine turned up as a small messenger-group with 5 people, which is now a full-fledged community with ~150 members who moved to Germany.\\nNot onlysharing helps you to structure your knowledge, it bring people together, and we are nothing without people.\\n\\nSo, let's talk, we will never know where this call can bring us until we have it!\"}]"

	userContent := "Find 3 best mentors for this user by his request. Here user request:\n\"" + request + "\nIn response write ONLY list of mentors IDs. Dont include explanation or any other text"

	client := openai.NewClient(ApiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: OpenApiModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemContent,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userContent,
				},
			},
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return nil, err
	}

	log.Printf("chatgpt response %s\n", resp.Choices[0].Message.Content)
	mentorsIds := parseChatgptResponse(resp)

	filteredMentors := getMentorsFilteredWithChatgpt(mentorsForRequest, mentorsIds)

	go database.SaveBestMentorsForUser(userId, mentorsIds)

	return filteredMentors, nil
}

func getMentorsFilteredWithChatgpt(mentorsForRequest []model.MentorForRequest, mentorsIds []string) []model.MentorForRequest {
	filteredMentors := []model.MentorForRequest{}
	for _, mentor := range mentorsForRequest {
		for _, id := range mentorsIds {
			if mentor.MentorId == id {
				imageResult, err := database.GetUserPictureByUserId(id)
				if err == nil {
					mentor.UserImage.UserId = imageResult.UserId
					mentor.UserImage.Image = imageResult.Image
					mentor.UserImage.Extension = imageResult.Extension
				}

				filteredMentors = append(filteredMentors, mentor)
				break
			}
		}
	}
	return filteredMentors
}

func parseChatgptResponse(resp openai.ChatCompletionResponse) []string {
	cleanedString := strings.ReplaceAll(resp.Choices[0].Message.Content, `"`, "")
	cleanedString = strings.ReplaceAll(cleanedString, "[", "")
	cleanedString = strings.ReplaceAll(cleanedString, "]", "")
	stringArray := strings.Split(cleanedString, ", ")
	return stringArray
}
