package httpHandlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"oysterProject/database"
	"oysterProject/model"
)

func CreateSessionReview(w http.ResponseWriter, r *http.Request) {
	sessionId := chi.URLParam(r, "sessionId")
	sessionChan := make(chan *model.Session)
	errChan := make(chan error)
	go func() {
		session, err := database.GetMentorMenteeIdsBySessionId(sessionId)
		if err != nil {
			errChan <- err
			return
		}
		sessionChan <- session
	}()

	var sessionReview model.Review
	err := parseJSONRequest(r, &sessionReview)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON session review")
		return
	}

	var session *model.Session
	select {
	case sessionFromChan := <-sessionChan:
		session = sessionFromChan
	case errFromChan := <-errChan:
		if errFromChan != nil {
			writeMessageResponse(w, r, http.StatusNotFound, "Session not found")
			return
		}
	}
	sessionReview.FillDefaultsSessionReview(session)
	err = database.CreateReview(&sessionReview)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database error creating review")
		return
	}
	writeJSONResponse(w, r, http.StatusCreated, sessionReview)
}

func CreatePublicReview(w http.ResponseWriter, r *http.Request) {
	var sessionReview model.Review
	err := parseJSONRequest(r, &sessionReview)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON session review")
		return
	}
	sessionReview.FillDefaultsMentorReview()
	err = database.CreateReview(&sessionReview)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database error creating review")
		return
	}
	writeJSONResponse(w, r, http.StatusCreated, sessionReview)
}
