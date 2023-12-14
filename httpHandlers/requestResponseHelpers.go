package httpHandlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func ParseJSONRequest(r *http.Request, payload interface{}) error {
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		log.Printf("Error parsing JSON request error(%s), body(%s):\n", err, r.Body)
	}
	return err
}

func WriteMessageResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	writeResponse(w, message)
}

func WriteJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding JSON response:%v", err)
	}
}

func writeResponse(w http.ResponseWriter, message string) {
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Printf("Error writing response:%v\n", err)
	}
}

func handleInvalidTokenResponse(w http.ResponseWriter) {
	WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
}
