package httpHandlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func ParseJSONRequest(r *http.Request, payload interface{}) error {
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		log.Println(err)
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
		log.Println("Error encoding JSON response:", err)
	}
}

func writeResponse(w http.ResponseWriter, message string) {
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Println(err)
	}
}
