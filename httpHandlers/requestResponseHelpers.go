package httpHandlers

import (
	"encoding/json"
	"github.com/go-chi/render"
	"log"
	"net/http"
	"time"
)

func parseJSONRequest(r *http.Request, payload interface{}) error {
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		log.Printf("Error parsing JSON request error(%s), body(%s):\n", err, r.Body)
	}
	return err
}

func writeMessageResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	render.Status(r, status)
	render.PlainText(w, r, message)
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, status int, payload interface{}) {
	render.Status(r, status)
	render.JSON(w, r, payload)
}

func writeSessionCookie(w http.ResponseWriter, name, value string, time time.Time) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}
	http.SetCookie(w, cookie)
}

func deleteCookie(w http.ResponseWriter, name string) {
	cookie := http.Cookie{
		Name:    name,
		MaxAge:  -1,
		Path:    "/",
		Expires: time.Now().Add(-expirationTime),
	}
	http.SetCookie(w, &cookie)
}
