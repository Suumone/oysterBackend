package httpHandlers

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"oysterProject/database"
	"oysterProject/emailNotifications"
	"oysterProject/model"
	"oysterProject/utils"
	"strconv"
	"strings"
	"time"
)

const minimumTimeBetweenSessions = 30 * time.Minute
const sessionDuration = 60 * time.Minute

func GetUserAvailableWeekdays(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	userId := queryParameters.Get("id")
	userIdObj, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Invalid id")
		return
	}
	startDate, err := parseDateParameter(queryParameters.Get("from"))
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing date from")
		return
	}

	endDate, err := parseDateParameter(queryParameters.Get("to"))
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing date to")
		return
	}

	user, err := database.GetUserWithImageByID(userIdObj)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "User not found")
		return
	}

	if utils.IsEmptyStruct(user.Availability) {
		writeMessageResponse(w, r, http.StatusNotFound, "User does not have available timeslots")
		return
	}

	result := calculateAvailableWeekdays(user.Availability, startDate, endDate)
	writeJSONResponse(w, r, http.StatusOK, result)
}

func calculateAvailableWeekdays(availabilities []*model.Availability, startDate, endDate time.Time) []model.AvailableWeekday {
	var result []model.AvailableWeekday

	uniqueWeekdays := getUniqueWeekdays(availabilities)
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		if slices.Contains(uniqueWeekdays, currentDate.Weekday()) {
			result = append(result, model.AvailableWeekday{Date: currentDate, Weekday: currentDate.Weekday().String()})
		}

		currentDate = currentDate.Add(24 * time.Hour)
	}

	return result
}

func getUniqueWeekdays(availabilities []*model.Availability) []time.Weekday {
	uniqueWeekdays := make(map[string]struct{})
	for _, availability := range availabilities {
		uniqueWeekdays[availability.Weekday] = struct{}{}
	}
	var result []time.Weekday
	for weekday := range uniqueWeekdays {
		result = append(result, utils.GetDayOfWeek(weekday))
	}
	return result
}

func GetUserAvailableSlots(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	userId, err := primitive.ObjectIDFromHex(queryParameters.Get("id"))
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Invalid user id")
		return
	}
	startDate, err := parseDateParameter(queryParameters.Get("date"))
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing date")
		return
	}
	user, err := database.GetUserByID(userId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "User not found")
		return
	}
	bookedSessions, err := database.GetUserUpcomingSessions(user.Id, true) //todo in channel
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error user sessions info from database")
		return
	}

	endDate := time.Date(
		startDate.Year(),
		startDate.Month(),
		startDate.Day(),
		23, 59, 0, 0,
		startDate.Location(),
	)
	result := calculateAvailability(user.Availability, bookedSessions, startDate, endDate)
	writeJSONResponse(w, r, http.StatusOK, result)
}

func parseDateParameter(dateParam string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, dateParam)
	if err != nil {
		log.Printf("Error parsing date (%s): %v\n", dateParam, err)
		return time.Time{}, err
	}
	return date, nil
}

func calculateAvailability(availabilities []*model.Availability, bookedSessions []*model.SessionResponse, startDate, endDate time.Time) []model.TimeSlot {
	var result []model.TimeSlot

	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		for _, availability := range availabilities {
			if currentDate.Weekday() == utils.GetDayOfWeek(availability.Weekday) {
				slots := getSlots(availability, currentDate)
				result = append(result, excludeBookedSlots(slots, bookedSessions)...)
			}
		}

		currentDate = currentDate.Add(24 * time.Hour)
	}
	return result
}

func getSlots(availability *model.Availability, currentDate time.Time) []model.TimeSlot {
	var result []model.TimeSlot

	availabilityStart, availabilityEnd := getAvailabilityTimeRange(availability, currentDate)

	for current := availabilityStart; current.Before(availabilityEnd); current = current.Add(minimumTimeBetweenSessions) {
		if current.Before(availabilityEnd.Add(sessionDuration)) {
			result = append(result, model.TimeSlot{StartTime: current, EndTime: current.Add(sessionDuration)})
		}
	}
	return result
}

func excludeBookedSlots(slots []model.TimeSlot, bookedSessions []*model.SessionResponse) []model.TimeSlot {
	var availableSlots []model.TimeSlot

	for _, slot := range slots {
		isBooked := false
		for _, bookedSlot := range bookedSessions {
			if bookedSlot.SessionTimeStart != nil && slot.EndTime.After(*bookedSlot.SessionTimeStart) && bookedSlot.SessionTimeEnd != nil && slot.StartTime.Before(*bookedSlot.SessionTimeEnd) && (bookedSlot.SessionStatus < 5) {
				isBooked = true
				break
			}
		}

		if !isBooked {
			availableSlots = append(availableSlots, slot)
		}
	}

	return availableSlots
}

func getAvailabilityTimeRange(availability *model.Availability, currentDate time.Time) (time.Time, time.Time) {
	startTime := parseTime(availability.TimeFrom, currentDate)
	endTime := parseTime(availability.TimeTo, currentDate)
	return startTime, endTime
}

func parseTime(timeStr string, currentDate time.Time) time.Time {
	hours, minutes := parseHoursAndMinutes(timeStr)
	return time.Date(
		currentDate.Year(),
		currentDate.Month(),
		currentDate.Day(),
		hours, minutes,
		0, 0,
		currentDate.Location(),
	)
}

func parseHoursAndMinutes(timeStr string) (int, int) {
	parts := strings.Split(timeStr, ":")
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	return hours, minutes
}

func CreateSession(w http.ResponseWriter, r *http.Request) {
	var mentorSession model.Session
	err := parseJSONRequest(r, &mentorSession)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from session create request")
		return
	}
	err = setSessionDetails(&mentorSession)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "Mentor was not found in database: "+err.Error())
		return
	}
	updatedSession, err := database.CreateSession(mentorSession)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database session insert error: "+err.Error())
		return
	}
	go emailNotifications.SendSessionWasCreatedEmail(updatedSession)
	writeJSONResponse(w, r, http.StatusCreated, updatedSession)
}

func setSessionDetails(session *model.Session) error {
	mentor, err := database.GetUserByID(session.MentorId)
	if err != nil {
		log.Printf("CancelSession: Failed to find user(%s) err: %v\n", session.MentorId.Hex(), err)
		return err
	}
	session.MeetingLink = mentor.MeetingLink
	if mentor.Prices != nil {
		session.PaymentDetails = mentor.Prices[0].Price
	} else {
		session.PaymentDetails = "free"
	}
	session.SessionStatus = model.PendingByMentor
	sessionTimeEnd := (*session.SessionTimeStart).Add(60 * time.Minute)
	session.SessionTimeEnd = &sessionTimeEnd
	return nil
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	id := queryParameters.Get("id")
	mentorSession, err := database.GetSession(id)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "Session not found")
		return
	}
	writeJSONResponse(w, r, http.StatusCreated, mentorSession)
}

func GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	user, err := database.GetUserByID(userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "User not found")
		return
	}

	sessions, err := database.GetUserSessions(user.Id, user.AsMentor)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Error during search session")
		return
	}
	sessionsResponse := groupSessionsByStatus(sessions)
	writeJSONResponse(w, r, http.StatusOK, sessionsResponse)
}

func groupSessionsByStatus(sessions []*model.SessionResponse) model.GroupedSessions {
	groupedSessions := model.GroupedSessions{}
	for _, s := range sessions {
		switch {
		case s.SessionStatus < model.Confirmed:
			groupedSessions.PendingSessions = append(groupedSessions.PendingSessions, s)
		case s.SessionStatus == model.Confirmed:
			groupedSessions.UpcomingSessions = append(groupedSessions.UpcomingSessions, s)
		case s.SessionStatus > model.Confirmed:
			groupedSessions.PastSessions = append(groupedSessions.PastSessions, s)
		}
	}
	return groupedSessions
}

func RescheduleRequest(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}

	var mentorSession model.Session
	err := parseJSONRequest(r, &mentorSession)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from session reschedule request")
		return
	}
	if mentorSession.NewSessionTimeStart == nil || mentorSession.SessionId.IsZero() {
		writeMessageResponse(w, r, http.StatusBadRequest, "Invalid json request")
		return
	}
	user, err := database.GetUserByID(userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusNotFound, "Failed to find user(%s)")
		return
	}
	sessionTimeEnd := (*mentorSession.NewSessionTimeStart).Add(60 * time.Minute)
	mentorSession.NewSessionTimeEnd = &sessionTimeEnd
	setRescheduleStatus(&mentorSession, user.AsMentor)
	updatedSession, err := database.RescheduleSession(mentorSession)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database error during session update")
	}

	go emailNotifications.SendSessionRescheduledEmail(updatedSession)
	writeJSONResponse(w, r, http.StatusOK, updatedSession)
}

func setRescheduleStatus(session *model.Session, isMentor bool) {
	if isMentor {
		session.SessionStatus = model.ReschedulingByMentor
	} else {
		session.SessionStatus = model.ReschedulingByMentee
	}
}

func ConfirmRescheduleRequest(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	sessionId := queryParameters.Get("sessionId")
	if sessionId == "" {
		writeMessageResponse(w, r, http.StatusBadRequest, "Session id wasn't provided")
		return
	}
	updatedSession, err := database.ConfirmSession(sessionId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database error during session confirm")
		return
	}
	go emailNotifications.SendSessionConfirmedEmail(updatedSession)
	writeJSONResponse(w, r, http.StatusOK, updatedSession)
}

func CancelRescheduleRequest(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	queryParameters := r.URL.Query()
	sessionId := queryParameters.Get("sessionId")
	sessionIdObj, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Session id invalid")
		return
	}
	updateSession, err := database.CancelSession(sessionIdObj, userSession.UserId)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database error during session cancel")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, updateSession)
}

func CreateSessionReview(w http.ResponseWriter, r *http.Request) {
	var sessionReview model.SessionReview
	err := parseJSONRequest(r, &sessionReview)
	if err != nil {
		writeMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON session review")
		return
	}

	mentorSession, err := database.CreateReviewAndUpdateSession(&sessionReview)
	if err != nil {
		writeMessageResponse(w, r, http.StatusInternalServerError, "Database error creating review")
		return
	}
	writeJSONResponse(w, r, http.StatusCreated, mentorSession)
}
