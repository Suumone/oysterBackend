package httpHandlers

import (
	"errors"
	"github.com/golang-jwt/jwt"
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
	users, err := database.GetMentors(queryParameters)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			WriteJSONResponse(w, http.StatusBadRequest, "Error parsing offset and limit")
			return
		} else {
			WriteJSONResponse(w, http.StatusInternalServerError, "Error getting mentors from database")
			return
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
	userId, err := getUserIdFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}

	user := database.GetUserByID(userId)
	if utils.IsEmptyStruct(user) {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, user)
}

func UpdateProfileByToken(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}

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

func getUserIdFromRequest(r *http.Request) (string, error) {
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
	}
	WriteJSONResponse(w, http.StatusOK, users)
}

func GetCurrentState(w http.ResponseWriter, r *http.Request) {
	userId, ok := getUserIdFromToken(w, r)
	if !ok {
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
	userId, ok := getUserIdFromToken(w, r)
	if !ok {
		return
	}
	var userForStateUpdate model.UserState
	if err := ParseJSONRequest(r, &userForStateUpdate); err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	if err := database.UpdateUserState(userForStateUpdate.AsMentor, userId); err != nil {
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

func getUserIdFromToken(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return "", false
	}
	userId, _ := claims["id"].(string)
	return userId, true
}
