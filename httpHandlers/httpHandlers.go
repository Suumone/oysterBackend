package httpHandlers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
	"net/http"
	"oysterProject/database"
	"oysterProject/model"
	"oysterProject/utils"
	"strings"
)

func HandleCreateMentor(w http.ResponseWriter, r *http.Request) {
	var payload model.Users
	if err := ParseJSONRequest(w, r, &payload); err != nil {
		return
	}
	utils.NormalizeSocialLinks(&payload)

	insertedID, err := database.SaveMentorInDB(payload)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error saving user to MongoDB")
		return
	}
	WriteJSONResponse(w, http.StatusCreated, insertedID)
}

func HandleGetMentors(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	users := database.GetMentorsFromDB(queryParameters)
	WriteJSONResponse(w, http.StatusOK, users)
}

func HandleGetMentorByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	mentor := database.GetMentorByIDFromDB(id)
	if utils.IsEmptyStruct(mentor) {
		WriteMessageResponse(w, http.StatusNotFound, "Mentor not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, mentor)
}

func HandleGetMentorReviews(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userWithReviews := database.GetMentorReviewsByIDFromDB(id)
	if utils.IsEmptyStruct(userWithReviews) {
		WriteMessageResponse(w, http.StatusNotFound, "Reviews not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, userWithReviews)
}

func HandleUpdateMentor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload model.Users
	if err := ParseJSONRequest(w, r, &payload); err != nil {
		return
	}

	if err := database.UpdateMentorInDB(payload, id); err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}
	WriteJSONResponse(w, http.StatusOK, id)
}

func HandleGetProfileByToken(w http.ResponseWriter, r *http.Request) {
	tokenStr := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	userId, _ := claims["id"].(string)

	user := database.GetMentorByIDFromDB(userId)
	if utils.IsEmptyStruct(user) {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, user)
}
