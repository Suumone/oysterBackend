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
	"path/filepath"
	"strconv"
	"strings"
)

const imageLimitSizeMB = 5

var allowedExtensions = []string{".jpg", ".jpeg", ".png"}

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

func GetMentorListFilters(w http.ResponseWriter, _ *http.Request) {
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
	tokenStr := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]

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

func GetCurrentState(w http.ResponseWriter, r *http.Request) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	userId, _ := claims["id"].(string)

	userState := database.GetCurrentState(userId)
	if utils.IsEmptyStruct(userState) {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	WriteJSONResponse(w, http.StatusOK, userState)
}

func UpdateCurrentState(w http.ResponseWriter, r *http.Request) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	userId, _ := claims["id"].(string)
	var userForUpdate model.UserState
	if err := ParseJSONRequest(r, &userForUpdate); err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	if err := database.UpdateUserState(userForUpdate.AsMentor, userId); err != nil {
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

func UploadUserImage(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}

	err = r.ParseMultipartForm(1024 * 1024 * imageLimitSizeMB) // image size limit in mb
	if err != nil {
		log.Printf("Error parsing multipart form: %v\n", err)
		WriteJSONResponse(w, http.StatusBadRequest, "File too big")
		return
	}

	file, header, err := r.FormFile("profilePicture")
	if err != nil {
		log.Printf("Error retrieving the file: %v\n", err)
		WriteJSONResponse(w, http.StatusInternalServerError, "Error Retrieving the file")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !utils.Contains(allowedExtensions, ext) {
		WriteJSONResponse(w, http.StatusBadRequest, "File type not allowed")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		WriteJSONResponse(w, http.StatusInternalServerError, "Error reading file")
		return
	}
	err = database.SaveProfilePicture(userId, fileBytes, ext)
	if err != nil {
		log.Printf("Error during saving picture: %v\n", err)
		WriteMessageResponse(w, http.StatusBadRequest, "Error during saving picture")
		return
	}
	WriteJSONResponse(w, http.StatusOK, "Profile picture successfully updated")
}

func GetUserImage(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	var userId string
	if len(queryParameters) == 0 {
		id, err := getUserIdFromRequest(r)
		if err != nil {
			WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
			return
		}
		userId = id
	} else {
		userId = queryParameters.Get("id")
	}
	userImage, err := database.GetUserPictureByUserId(userId)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Error getting image from database")
		return
	}
	WriteJSONResponse(w, http.StatusOK, userImage)
}

func GetImageConfigurations(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"imageLimitSizeMB":  imageLimitSizeMB,
		"allowedExtensions": allowedExtensions,
	}
	WriteJSONResponse(w, http.StatusOK, response)
}
