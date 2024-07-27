package httpHandlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sashabaranov/go-openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"net/url"
	"os"
	"oysterProject/database"
	"oysterProject/model"
	"strings"
)

const (
	openApiModel           = openai.GPT3Dot5Turbo1106
	systemContentBeginning = "You will be provided with a list of mentors to analyze it. After this, the user will submit their request. Your task is to find THREE best mentor for him based on his request and the list of mentors. In response write ONLY list of  3 mentor IDs. Dont include explanation or any other text.\nField Explanation:\n \"mentorId\" - mentor id use it for response on my requests.\n\"areaOfExpertise\": The mentor's area of expertise.\n\"company\": The mentor's current employer.\n\"countryDescription\": A list of countries where the mentor has lived and worked, along with descriptions of their experiences in those countries.\n\"experience\": The mentor's total years of experience.\n\"industryExpertise\": A list of industries in which the mentor has expertise.\n\"jobTitle\": The mentor's job title.\n\"language\": The languages the mentor can communicate in.\n\"mentorsTopics\": A list of topics or areas in which the mentor can provide guidance.\n\"name\": The mentor's name.\n\"skill\": A list of skills the mentor possesses.\n\"welcomeText\": A brief introduction and welcome message from the mentor, including their background, experience and other information.\n\nList of mentors:\n"
)

var (
	ApiKey = os.Getenv("OPENAI_API_KEY")
)

func mapUserToMentorForRequest(user *model.User) model.MentorForRequest {
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
		if countryDesc.Country != "" && countryDesc.Description != "" {
			mentor.CountryDescription = append(mentor.CountryDescription, model.CountryDescription{
				Country:     countryDesc.Country,
				Description: countryDesc.Description,
			})
		}
	}

	for _, mentorTopic := range user.MentorsTopics {
		if mentorTopic.Description != "" && mentorTopic.Topic != "" {
			mentor.MentorsTopics = append(mentor.MentorsTopics, model.MentorsTopics{
				Description: mentorTopic.Description,
				Topic:       mentorTopic.Topic,
			})
		}
	}

	for _, areaOfExpertise := range user.AreaOfExpertise {
		if areaOfExpertise.Area != "" && areaOfExpertise.Experience != 0 {
			mentor.AreaOfExpertise = append(mentor.AreaOfExpertise, model.AreaOfExpertise{
				Area:       areaOfExpertise.Area,
				Experience: areaOfExpertise.Experience,
			})
		}
	}

	for _, price := range user.Prices {
		if price.Price != "" {
			mentor.Prices = append(mentor.Prices, model.Price{
				Price: price.Price,
			})
		}
	}

	return mentor
}

func CalculateBestMentors(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	var requestPayload model.ChatgptHttpPayload
	if err := parseJSONRequest(r, &requestPayload); err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	user, err := database.GetUserByID(userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "User not found")
		return
	}
	if user.UserMentorRequest == requestPayload.Request {
		mentors, err := database.GetBestMentors(userSession.UserId)
		if err != nil {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error searching for mentors in database")
			return
		}
		var mentorsResponse []model.MentorForRequest
		for _, mentor := range mentors {
			mentorsResponse = append(mentorsResponse, mapUserToMentorForRequest(mentor))
		}
		writeJSONResponse(w, r, http.StatusOK, mentors)
	} else {
		go database.UpdateMentorRequest(requestPayload.Request, userSession.UserId)

		mentorsFromChatgpt, err := SendRequestToChatgpt(requestPayload.Request, userSession.UserId)
		if err != nil {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error searching for mentors")
			return
		}
		writeJSONResponse(w, r, http.StatusOK, mentorsFromChatgpt)
	}
}

func SendRequestToChatgpt(request string, userId primitive.ObjectID) ([]model.MentorForRequest, error) {
	mentors, err := database.GetMentors(nil, primitive.NilObjectID, nil)
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
	systemContent := systemContentBeginning + jsonMentorsForRequestString
	client := openai.NewClient(ApiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openApiModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemContent,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: request,
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

	if len(mentorsIds) > 0 {
		filteredMentors := getMentorsFilteredWithChatgpt(mentorsForRequest, mentorsIds)
		go database.SaveBestMentorsForUser(userId, filteredMentors)
		return filteredMentors, nil
	} else {
		var result []model.MentorForRequest
		mentorsFromDb, err := database.GetMentors(url.Values{"offset": []string{"0"}, "limit": []string{"3"}}, primitive.NilObjectID, nil)
		if err != nil {
			return result, nil
		}
		for _, user := range mentorsFromDb {
			result = append(result, mapUserToMentorForRequest(user))
		}
		return result, nil
	}
}

func getMentorsFilteredWithChatgpt(mentorsForRequest []model.MentorForRequest, mentorsIds []string) []model.MentorForRequest {
	var mentorsIdsObj []primitive.ObjectID
	for _, id := range mentorsIds {
		idObj, _ := primitive.ObjectIDFromHex(id)
		mentorsIdsObj = append(mentorsIdsObj, idObj)
	}
	usersWithImages, err := database.GetUserImages(mentorsIdsObj)
	if err != nil {
		return nil
	}
	userImagesMap := make(map[string]*model.UserImage)
	for _, userImage := range usersWithImages {
		userImagesMap[userImage.UserId.Hex()] = userImage
	}

	var resultMentors []model.MentorForRequest
	for _, mentor := range mentorsForRequest {
		if userImage, ok := userImagesMap[mentor.MentorId]; ok {
			mentor.UserImage = userImage
			resultMentors = append(resultMentors, mentor)
		}
	}

	return resultMentors
}

func parseChatgptResponse(resp openai.ChatCompletionResponse) []string {
	cleanedString := strings.ReplaceAll(resp.Choices[0].Message.Content, `"`, "")
	cleanedString = strings.ReplaceAll(cleanedString, " ", "")
	cleanedString = strings.ReplaceAll(cleanedString, "[", "")
	cleanedString = strings.ReplaceAll(cleanedString, "]", "")
	cleanedString = strings.ReplaceAll(cleanedString, ",", "\n")
	stringArray := strings.Split(cleanedString, "\n")
	return removeEmptyStrings(stringArray)
}

func removeEmptyStrings(slice []string) []string {
	var result []string
	for _, s := range slice {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
