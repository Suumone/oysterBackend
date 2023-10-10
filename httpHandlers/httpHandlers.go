package httpHandlers

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"net/mail"
	"oysterProject/database"
	"oysterProject/model"
	"oysterProject/utils"
	"strings"
)

func GetMentors(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	users := database.GetMentors(queryParameters)
	WriteJSONResponse(w, http.StatusOK, users)
}

func GetMentorListFilters(w http.ResponseWriter, r *http.Request) {
	listOfFilters, err := database.GetListOfFilterFields()
	if err != nil {
		log.Printf("Error getting fields filter: %v\n", err)
		WriteMessageResponse(w, http.StatusInternalServerError, "Error getting fields filter")
		return
	}

	WriteJSONResponse(w, http.StatusOK, listOfFilters)
}

func GetMentor(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	id := queryParameters.Get("id")

	mentor := database.GetUserByID(id)
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
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	userId, _ := claims["id"].(string)

	user := database.GetUserByID(userId)
	if utils.IsEmptyStruct(user) {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, user)
}

func UpdateProfileByToken(w http.ResponseWriter, r *http.Request) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	userId, _ := claims["id"].(string)

	var userForUpdate model.User
	if err := ParseJSONRequest(r, &userForUpdate); err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	_, err = mail.ParseAddress(userForUpdate.Email)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Email is not valid")
		return
	}
	utils.NormalizeSocialLinks(&userForUpdate)

	if err := database.UpdateUser(userForUpdate, userId); err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}
	WriteJSONResponse(w, http.StatusOK, "User updated")
}

func getTokenClaimsFromRequest(r *http.Request) (jwt.MapClaims, error) {
	tokenStr := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	return claims, err
}

func GetTopMentors(w http.ResponseWriter, r *http.Request) {
	users := database.GetTopMentors()
	WriteJSONResponse(w, http.StatusOK, users)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	userId, _ := claims["id"].(string)
	var passwordPayload model.PasswordChange
	err = ParseJSONRequest(r, &passwordPayload)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	err = database.ChangePassword(userId, passwordPayload)
	if err != nil {
		log.Printf("Error updating password: %v\n", err)
		WriteMessageResponse(w, http.StatusInternalServerError, "Error updating password")
		return
	}
	WriteJSONResponse(w, http.StatusOK, "Password successfully updated")
}
