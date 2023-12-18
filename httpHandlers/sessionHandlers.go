package httpHandlers

import (
	"golang.org/x/exp/slices"
	"log"
	"net/http"
	"oysterProject/database"
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
	startDate, err := parseDateParameter(queryParameters.Get("from"))
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing date from")
		return
	}

	endDate, err := parseDateParameter(queryParameters.Get("to"))
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing date to")
		return
	}

	user, err := getUserByID(userId)
	if err != nil {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}

	if utils.IsEmptyStruct(user.Availability) {
		WriteMessageResponse(w, http.StatusNotFound, "User does not have available timeslots")
		return
	}

	result := calculateAvailableWeekdays(user.Availability, startDate, endDate)
	WriteJSONResponse(w, http.StatusOK, result)
}

func calculateAvailableWeekdays(availabilities []model.Availability, startDate, endDate time.Time) []model.AvailableWeekday {
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

func getUniqueWeekdays(availabilities []model.Availability) []time.Weekday {
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
	userId := queryParameters.Get("id")
	startDate, err := parseDateParameter(queryParameters.Get("date"))
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing date")
		return
	}

	user, err := getUserByID(userId)
	if err != nil {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}
	bookedSessions, err := database.GetUserSessions(user.Id, true)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error user sessions info from database")
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
	WriteJSONResponse(w, http.StatusOK, result)
}

func parseDateParameter(dateParam string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, dateParam)
	if err != nil {
		log.Printf("Error parsing date (%s): %v\n", dateParam, err)
		return time.Time{}, err
	}
	return date, nil
}

func calculateAvailability(availabilities []model.Availability, bookedSessions []model.SessionResponse, startDate, endDate time.Time) []model.TimeSlot {
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

func getSlots(availability model.Availability, currentDate time.Time) []model.TimeSlot {
	var result []model.TimeSlot

	availabilityStart, availabilityEnd := getAvailabilityTimeRange(availability, currentDate)

	for current := availabilityStart; current.Before(availabilityEnd); current = current.Add(minimumTimeBetweenSessions) {
		if current.Before(availabilityEnd.Add(sessionDuration)) {
			result = append(result, model.TimeSlot{StartTime: current, EndTime: current.Add(sessionDuration)})
		}
	}
	return result
}

func excludeBookedSlots(slots []model.TimeSlot, bookedSessions []model.SessionResponse) []model.TimeSlot {
	var availableSlots []model.TimeSlot

	for _, slot := range slots {
		isBooked := false
		for _, bookedSlot := range bookedSessions {
			if slot.EndTime.After(*bookedSlot.SessionTimeStart) && slot.StartTime.Before(*bookedSlot.SessionTimeEnd) && (bookedSlot.SessionStatus < 5) {
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

func getAvailabilityTimeRange(availability model.Availability, currentDate time.Time) (time.Time, time.Time) {
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
	var session model.Session
	err := ParseJSONRequest(r, &session)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from session create request")
		return
	}
	setSessionDetails(&session)
	updatedSession, err := database.CreateSession(session)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Database session insert error: "+err.Error())
		return
	}
	WriteJSONResponse(w, http.StatusCreated, updatedSession)
}

func setSessionDetails(session *model.Session) {
	mentor := database.GetUserByID(session.MentorId.Hex())
	session.MeetingLink = mentor.MeetingLink
	if mentor.Prices != nil {
		session.PaymentDetails = mentor.Prices[0].Price
	}
	session.SessionStatus = model.PendingByMentor
	sessionTimeEnd := (*session.SessionTimeStart).Add(60 * time.Minute)
	session.SessionTimeEnd = &sessionTimeEnd
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	queryParameters := r.URL.Query()
	id := queryParameters.Get("id")
	session, err := database.GetSession(id)
	if err != nil {
		WriteJSONResponse(w, http.StatusNotFound, "Session not found")
		return
	}
	WriteJSONResponse(w, http.StatusCreated, session)
}

func GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}
	user, err := getUserByID(userId)
	if err != nil {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}

	sessions, err := database.GetUserSessions(user.Id, user.AsMentor)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error during search session")
		return
	}
	sessionsResponse := groupSessionsByStatus(sessions)
	WriteJSONResponse(w, http.StatusOK, sessionsResponse)
}

func groupSessionsByStatus(sessions []model.SessionResponse) model.GroupedSessions {
	groupedSessions := model.GroupedSessions{}
	for _, session := range sessions {
		switch {
		case session.SessionStatus < model.Confirmed:
			groupedSessions.PendingSessions = append(groupedSessions.PendingSessions, session)
		case session.SessionStatus == model.Confirmed:
			groupedSessions.UpcomingSessions = append(groupedSessions.UpcomingSessions, session)
		case session.SessionStatus > model.Confirmed:
			groupedSessions.UpcomingSessions = append(groupedSessions.UpcomingSessions, session)
		}
	}
	return groupedSessions
}

func RescheduleRequest(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}

	var session model.Session
	err = ParseJSONRequest(r, &session)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from session reschedule request")
		return
	}
	user := database.GetUserByID(userId)
	sessionTimeEnd := (*session.NewSessionTimeStart).Add(60 * time.Minute)
	session.NewSessionTimeEnd = &sessionTimeEnd
	setRescheduleStatus(&session, user.AsMentor)
	updatedSession, err := database.RescheduleSession(session)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Database error during session update")
	}
	WriteJSONResponse(w, http.StatusOK, updatedSession)
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
		WriteMessageResponse(w, http.StatusBadRequest, "Session id wasn't provided")
		return
	}
	updateSession, err := database.ConfirmSession(sessionId)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Database error during session confirm")
		return
	}
	WriteJSONResponse(w, http.StatusOK, updateSession)
}

func CancelRescheduleRequest(w http.ResponseWriter, r *http.Request) {
	userId, err := getUserIdFromToken(r)
	if err != nil {
		handleInvalidTokenResponse(w)
		return
	}
	queryParameters := r.URL.Query()
	sessionId := queryParameters.Get("sessionId")
	if sessionId == "" {
		WriteMessageResponse(w, http.StatusBadRequest, "Session id wasn't provided")
		return
	}
	updateSession, err := database.CancelSession(sessionId, userId)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Database error during session cancel")
		return
	}
	WriteJSONResponse(w, http.StatusOK, updateSession)
}
