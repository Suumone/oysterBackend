package httpHandlers

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"net/mail"
	"oysterProject/database"
	"oysterProject/emailNotifications"
	"oysterProject/model"
	"oysterProject/utils"
	"strconv"
)

func GetMentorsList(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	users, err := database.GetMentors(queryParameters, userSession.UserId)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing offset and limit")
			return
		} else {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting mentors from database")
			return
		}
	}

	var usersIdsObj []primitive.ObjectID
	for _, user := range users {
		usersIdsObj = append(usersIdsObj, user.Id)
	}
	usersWithImages, err := database.GetUserImages(usersIdsObj)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting image from database for mentors")
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

	writeJSONResponse(w, r, http.StatusOK, users)
}

func GetMentorListFilters(w http.ResponseWriter, r *http.Request) {
	var listOfFilters []map[string]interface{}
	var requestParams model.RequestParams
	err := parseJSONRequest(r, &requestParams)
	if err == nil {
		listOfFilters, err = database.GetFiltersByNames(requestParams)
		if err != nil {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting fields filter")
			return
		}
	} else if err == io.EOF {
		listOfFilters, err = database.GetListOfFilterFields()
		if err != nil {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting fields filter")
			return
		}
	} else {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, listOfFilters)
}

func GetMentor(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	id := queryParameters.Get("id")
	idToFind, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("GetMentor: error converting id to objectId: %v\n", err)
		writeMessageResponse(w, r, http.StatusBadRequest, "Invalid id")
	}

	mentor, err := database.GetUserWithImageByID(idToFind)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "Mentor not found")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, mentor)
}

func GetMentorReviews(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	mentorId := queryParameters.Get("mentorId")
	if len(mentorId) > 0 {
		userWithReviews, err := database.GetMentorReviewsByID(mentorId)
		if err != nil {
			writeMessageResponse(w, r, http.StatusNotFound, "Reviews not found")
			return
		}
		writeJSONResponse(w, r, http.StatusOK, userWithReviews)
	} else {
		reviews, err := database.GetReviewsForFrontPage()
		if err != nil {
			writeMessageResponse(w, r, http.StatusNotFound, "Reviews not found")
			return
		}

		writeJSONResponse(w, r, http.StatusOK, reviews)
	}
}

func GetProfileByToken(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}

	user, err := database.GetUserWithImageByID(userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "User not found")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, user)
}

func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}

	var userForUpdate model.User
	if err := parseJSONRequest(r, &userForUpdate); err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	if userForUpdate.Email != "" {
		_, err := mail.ParseAddress(userForUpdate.Email)
		if err != nil {
			writeMessageResponse(w, r, http.StatusBadRequest, "Email is not valid")
			return
		}
	}

	userAfterUpdate, err := database.UpdateAndGetUser(&userForUpdate, userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}
	userForExperienceUpdate := &model.User{}
	for _, entry := range userAfterUpdate.AreaOfExpertise {
		userForExperienceUpdate.Experience += entry.Experience
	}
	if userAfterUpdate.IsNewUser == true {
		userForExperienceUpdate.IsNewUser = false

		go emailNotifications.SendUserFilledQuestionsEmail(userAfterUpdate)
	}
	userForExperienceUpdate, err = database.UpdateAndGetUser(userForExperienceUpdate, userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}

	writeJSONResponse(w, r, http.StatusOK, userForExperienceUpdate)
}

func GetTopMentors(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	users, err := database.GetTopMentors(queryParameters)
	if err != nil {
		if errors.Is(err, strconv.ErrSyntax) {
			writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing offset and limit")
			return
		} else {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting mentors from database")
			return
		}
	} else if len(users) == 0 {
		writeJSONResponse(w, r, http.StatusNotFound, users)
		return
	}

	var usersIdsObj []primitive.ObjectID
	for _, user := range users {
		usersIdsObj = append(usersIdsObj, user.Id)
	}
	usersWithImages, err := database.GetUserImages(usersIdsObj)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting images from database for users")
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
	writeJSONResponse(w, r, http.StatusOK, users)
}

func GetCurrentState(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	userState, err := database.GetCurrentState(userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "User not found")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, userState)
}

func UpdateCurrentState(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	if err := database.UpdateUserState(userSession.UserId); err != nil {
		if errors.Is(err, utils.UserIsNotMentor) {
			writeMessageResponse(w, r, http.StatusBadRequest, "Status update for mentors only")
			return
		} else {
			writeMessageResponse(w, r, http.StatusInternalServerError, "Error updating user to MongoDB")
			return
		}
	}
	writeMessageResponse(w, r, http.StatusOK, "User state updated")
}

func GetListValues(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	listOfValues, err := database.GetValuesForSelect(queryParameters)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error reading values from database")
	}
	writeJSONResponse(w, r, http.StatusOK, listOfValues)
}
