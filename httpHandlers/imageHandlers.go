package httpHandlers

import (
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
	userId, err := getUserIdFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	err = r.ParseMultipartForm(utils.ImageLimitSizeMB)
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
		"imageLimitSizeMB":  utils.ImageLimitSizeMB / (1024 * 1024),
		"allowedExtensions": allowedExtensions,
	}
	WriteJSONResponse(w, http.StatusOK, response)
}
