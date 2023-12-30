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
	/*
		render.Status(r, status)
		render.JSON(w, r, payload)
	*/
}

func writeResponse(w http.ResponseWriter, message string) {
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Printf("Error writing response:%v\n", err)
	}
}

func handleInvalidTokenResponse(w http.ResponseWriter) {
	WriteMessageResponse(w, http.StatusForbidden, "Invalid token")
}

// todo
//func WriteSessionCookie(w http.ResponseWriter, token string, expiry time.Time) {
//	cookie := &http.Cookie{
//		Name:     s.Cookie.Name,
//		Value:    token,
//		Path:     s.Cookie.Path,
//		Domain:   s.Cookie.Domain,
//		Secure:   s.Cookie.Secure,
//		HttpOnly: s.Cookie.HttpOnly,
//		SameSite: s.Cookie.SameSite,
//	}
//
//	if expiry.IsZero() {
//		cookie.Expires = time.Unix(1, 0)
//		cookie.MaxAge = -1
//	} else if s.Cookie.Persist || s.GetBool(ctx, "__rememberMe") {
//		cookie.Expires = time.Unix(expiry.Unix()+1, 0)        // Round up to the nearest second.
//		cookie.MaxAge = int(time.Until(expiry).Seconds() + 1) // Round up to the nearest second.
//	}
//
//	w.Header().Add("Set-Cookie", cookie.String())
//	w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)
//}
