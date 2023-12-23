package httpHandlers

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"net/mail"
	"oysterProject/database"
	"oysterProject/model"
	"oysterProject/utils"
	"strconv"
	"strings"
)

func GetMentorsList(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}
	users, err := database.GetMentors(queryParameters, userId)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			WriteJSONResponse(w, http.StatusBadRequest, "Error parsing offset and limit")
			return
		} else {
			WriteJSONResponse(w, http.StatusInternalServerError, "Error getting mentors from database")
			return
		}
	}

	var usersIdsObj []primitive.ObjectID
	for _, user := range users {
		usersIdsObj = append(usersIdsObj, user.Id)
	}
	usersWithImages, err := database.GetUsersWithImages(usersIdsObj)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Error getting image from database for mentors")
		return
	}
	userImagesMap := make(map[primitive.ObjectID]*model.UserImage)
	for _, userImage := range usersWithImages {
		userImagesMap[userImage.UserId] = userImage
	}
	for i, user := range users {
		if userImage, ok := userImagesMap[user.Id]; ok {
			users[i].UserImage = userImage
		}
	}

	WriteJSONResponse(w, http.StatusOK, users)
}

func GetMentorListFilters(w http.ResponseWriter, r *http.Request) {
	var listOfFilters []map[string]interface{}
	var requestParams model.RequestParams
	err := ParseJSONRequest(r, &requestParams)
	if err == nil {
		listOfFilters, err = database.GetFiltersByNames(requestParams)
		if err != nil {
			log.Printf("Error getting fields filter: %v\n", err)
			WriteMessageResponse(w, http.StatusInternalServerError, "Error getting fields filter")
			return
		}
	} else if err == io.EOF {
		listOfFilters, err = database.GetListOfFilterFields()
		if err != nil {
			log.Printf("Error getting fields filter: %v\n", err)
			WriteMessageResponse(w, http.StatusInternalServerError, "Error getting fields filter")
			return
		}
	} else {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	WriteJSONResponse(w, http.StatusOK, listOfFilters)
}

func GetMentor(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	id := queryParameters.Get("id")

	mentor := database.GetUserWithImageByID(id)
	if utils.IsEmptyStruct(mentor) {
		WriteMessageResponse(w, http.StatusNotFound, "Mentor not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, mentor)
}

func GetMentorReviews(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	mentorId := queryParameters.Get("mentorId")
	if len(mentorId) > 0 {
		userWithReviews := database.GetMentorReviewsByID(mentorId)
		if utils.IsEmptyStruct(userWithReviews) {
			WriteMessageResponse(w, http.StatusNotFound, "Reviews not found")
			return
		}
		WriteJSONResponse(w, http.StatusOK, userWithReviews)
	} else {
		reviews := database.GetReviewsForFrontPage()
		if utils.IsEmptyStruct(reviews) {
			WriteMessageResponse(w, http.StatusNotFound, "Reviews not found")
			return
		}

		WriteJSONResponse(w, http.StatusOK, reviews)
	}
}

func GetProfileByToken(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}

	user := database.GetUserWithImageByID(userId)
	if utils.IsEmptyStruct(user) {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, user)
}

func getUserByID(userId string) (model.User, error) {
	user := database.GetUserWithImageByID(userId)
	if utils.IsEmptyStruct(user) {
		return model.User{}, errors.New("user not found")
	}
	return user, nil
}

func UpdateProfileByToken(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}

	var userForUpdate model.User
	if err := ParseJSONRequest(r, &userForUpdate); err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	if userForUpdate.Email != "" {
		_, err = mail.ParseAddress(userForUpdate.Email)
		if err != nil {
			WriteMessageResponse(w, http.StatusBadRequest, "Email is not valid")
			return
		}
	}
	//utils.NormalizeSocialLinks(&userForUpdate)

	userAfterUpdate, err := database.UpdateUser(userForUpdate, userId)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}
	var userForExperienceUpdate model.User
	for _, entry := range userAfterUpdate.AreaOfExpertise {
		userForExperienceUpdate.Experience += entry.Experience
	}
	userForExperienceUpdate, err = database.UpdateUser(userForExperienceUpdate, userId)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}

	WriteJSONResponse(w, http.StatusOK, userForExperienceUpdate)
}

func getUserIdFromToken(r *http.Request) (string, error) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		return "", err
	}
	userId, _ := claims["id"].(string)
	return userId, nil
}

func getTokenClaimsFromRequest(r *http.Request) (jwt.MapClaims, error) {
	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		return nil, errors.New("received empty token")
	}
	tokenStr := strings.Split(authToken, "Bearer ")[1]

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	return claims, err
}

func GetTopMentors(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	users, err := database.GetTopMentors(queryParameters)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			WriteJSONResponse(w, http.StatusBadRequest, "Error parsing offset and limit")
			return
		} else {
			WriteJSONResponse(w, http.StatusInternalServerError, "Error getting mentors from database")
			return
		}
	} else if len(users) == 0 {
		WriteJSONResponse(w, http.StatusNotFound, users)
		return
	}

	var usersIdsObj []primitive.ObjectID
	for _, user := range users {
		usersIdsObj = append(usersIdsObj, user.Id)
	}
	usersWithImages, err := database.GetUsersWithImages(usersIdsObj)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Error getting images from database for users")
		return
	}
	userImagesMap := make(map[primitive.ObjectID]*model.UserImage)
	for _, userImage := range usersWithImages {
		userImagesMap[userImage.UserId] = userImage
	}

	for i, user := range users {
		if userImage, ok := userImagesMap[user.Id]; ok {
			users[i].UserImage = userImage
		}
	}
	WriteJSONResponse(w, http.StatusOK, users)
}

func GetCurrentState(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}
	userState := database.GetCurrentState(userId)
	if utils.IsEmptyStruct(userState) {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, userState)
}

func UpdateCurrentState(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}
	if err := database.UpdateUserState(userId); err != nil {
		if errors.Is(err, utils.UserIsNotMentor) {
			WriteMessageResponse(w, http.StatusBadRequest, "Status update for mentors only")
			return
		} else {
			WriteMessageResponse(w, http.StatusInternalServerError, "Error updating user to MongoDB")
			return
		}
	}
	WriteJSONResponse(w, http.StatusOK, "User state updated")
}

func GetListValues(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	listOfValues, err := database.GetValuesForSelect(queryParameters)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error reading values from database")
	}
	WriteJSONResponse(w, http.StatusOK, listOfValues)
}
