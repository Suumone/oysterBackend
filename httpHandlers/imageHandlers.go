package httpHandlers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"log"
	"net/http"
	"oysterProject/database"
	"oysterProject/utils"
	"path/filepath"
	"strings"
)

var allowedExtensions = []string{".jpg", ".jpeg", ".png", ".heic"}

func UploadUserImage(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	err := r.ParseMultipartForm(utils.ImageLimitSizeMB)
	if err != nil {
		log.Printf("Error parsing multipart form: %v\n", err)
		writeMessageResponse(w, r, http.StatusBadRequest, "File too big")
		return
	}

	file, header, err := r.FormFile("profilePicture")
	if err != nil {
		log.Printf("Error retrieving the file: %v\n", err)
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error Retrieving the file")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !utils.Contains(allowedExtensions, ext) {
		writeMessageResponse(w, r, http.StatusBadRequest, "File type not allowed")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error reading file")
		return
	}
	err = database.SaveProfilePicture(userSession.UserId, fileBytes, ext)
	if err != nil {
		log.Printf("Error during saving picture: %v\n", err)
		writeMessageResponse(w, r, http.StatusBadRequest, "Error during saving picture")
		return
	}
	writeMessageResponse(w, r, http.StatusOK, "Profile picture successfully updated")
}

func GetUserImage(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	var userId primitive.ObjectID
	if len(queryParameters) == 0 {
		userSession := getUserSessionFromRequest(r)
		if userSession == nil {
			writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
			return
		}
		userId = userSession.UserId
	} else {
		var err error
		userId, err = primitive.ObjectIDFromHex(queryParameters.Get("id"))
		if err != nil {
			log.Printf("GetUserImage: error converting id to objectId: %v\n", err)
			writeMessageResponse(w, r, http.StatusBadRequest, "Invalid id")
		}
	}
	userImage, err := database.GetUserPictureByUserId(userId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error getting image from database")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, userImage)
}

func GetImageConfigurations(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"imageLimitSizeMB":  utils.ImageLimitSizeMB / (1024 * 1024),
		"allowedExtensions": allowedExtensions,
	}
	writeJSONResponse(w, r, http.StatusOK, response)
}
