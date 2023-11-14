package httpHandlers

import (
	"errors"
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

func calculateAvailableWeekdays(availabilities []model.Availability, startDate, endDate time.Time) []AvailableWeekday {
	var result []AvailableWeekday

	uniqueWeekdays := getUniqueWeekdays(availabilities)
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		if slices.Contains(uniqueWeekdays, currentDate.Weekday()) {
			result = append(result, AvailableWeekday{Date: currentDate, Weekday: currentDate.Weekday().String()})
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

	endDate := time.Date(
		startDate.Year(),
		startDate.Month(),
		startDate.Day(),
		23, 59, 0, 0,
		startDate.Location(),
	)
	result := calculateAvailability(user.Availability, startDate, endDate)
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

func getUserByID(userId string) (model.User, error) {
	user := database.GetUserByID(userId)
	if utils.IsEmptyStruct(user) {
		return model.User{}, errors.New("user not found")
	}
	return user, nil
}

func calculateAvailability(availabilities []model.Availability, startDate, endDate time.Time) []TimeSlot {
	var result []TimeSlot

	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		for _, availability := range availabilities {
			if currentDate.Weekday() == utils.GetDayOfWeek(availability.Weekday) {
				result = append(result, getSlots(availability, currentDate)...)
			}
		}

		currentDate = currentDate.Add(24 * time.Hour)
	}
	return result
}

func getSlots(availability model.Availability, currentDate time.Time) []TimeSlot {
	var result []TimeSlot

	availabilityStart, availabilityEnd := getAvailabilityTimeRange(availability, currentDate)

	for current := availabilityStart; current.Before(availabilityEnd); current = current.Add(minimumTimeBetweenSessions) {
		if current.Before(availabilityEnd.Add(sessionDuration)) {
			result = append(result, TimeSlot{StartTime: current, EndTime: current.Add(sessionDuration)})
		}
	}
	return result
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

type AvailableWeekday struct {
	Date    time.Time `json:"date"`
	Weekday string    `json:"weekday"`
}

type TimeSlot struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
