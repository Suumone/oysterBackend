package utils

import (
	"log"
	"net/url"
	"oysterProject/model"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	ImageLimitSizeMB = 1024 * 1024 * 5 //5 MB
	DateLayout       = "2006-01-02 15:04"
	TimeLayout       = "15:04"
)

func IsEmptyStruct(input interface{}) bool {
	zeroValue := reflect.New(reflect.TypeOf(input)).Elem().Interface()
	return reflect.DeepEqual(input, zeroValue)
}

func NormalizeSocialLinks(user *model.User) {
	user.LinkedInLink = makeURL(user.LinkedInLink, "linkedin.com/")
	user.InstagramLink = makeURL(user.InstagramLink, "instagram.com/")
	user.FacebookLink = makeURL(user.FacebookLink, "facebook.com/")
}

func makeURL(text string, urlPrefix string) string {
	if _, err := url.ParseRequestURI(text); err == nil {
		return text
	}
	if strings.HasPrefix(text, urlPrefix) {
		return "https://www." + text
	}

	return "https://www." + urlPrefix + strings.ReplaceAll(text, " ", "_")
}

func Contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func GetDayOfWeek(day string) time.Weekday {
	switch day {
	case "Mon":
		return time.Monday
	case "Tue":
		return time.Tuesday
	case "Wed":
		return time.Wednesday
	case "Thu":
		return time.Thursday
	case "Fri":
		return time.Friday
	case "Sat":
		return time.Saturday
	case "Sun":
		return time.Sunday
	default:
		return time.Monday
	}
}

func SetStatusText(session *model.Session) {
	session.StatusForMentee = session.SessionStatus.GetStatusForMentee()
	session.StatusForMentor = session.SessionStatus.GetStatusForMentor()
	session.Status = session.SessionStatus.String()
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func UpdateTimezoneTime(availability *model.Availability) error {
	timeZoneOffset := time.Duration(availability.TimeZone) * time.Minute
	fullDateTimeFrom := "2006-01-02 " + availability.TimeFrom
	fullDateTimeTo := "2006-01-02 " + availability.TimeTo
	parsedTimeFrom, err := time.Parse(DateLayout, fullDateTimeFrom)
	if err != nil {
		log.Printf("UpdateTimezoneTime: error parsedTimeFrom. TimeFrom: %s, error:: %v\n", availability.TimeFrom, err)
		return err
	}
	parsedTimeTo, err := time.Parse(DateLayout, fullDateTimeTo)
	if err != nil {
		log.Printf("UpdateTimezoneTime: error parsedTimeTo. TimeTo: %s, error:: %v\n", availability.TimeTo, err)
		return err
	}
	parsedTimeFrom = parsedTimeFrom.Add(timeZoneOffset)
	availability.TimeFrom = parsedTimeFrom.UTC().Format(TimeLayout)

	parsedTimeTo = parsedTimeTo.Add(timeZoneOffset)
	availability.TimeTo = parsedTimeTo.UTC().Format(TimeLayout)

	availability.TimeZone = -availability.TimeZone
	return nil
}

func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func GetSessionTime(session *model.SessionResponse) (string, string) {
	if session.SessionTimeStart != nil {
		return session.SessionTimeStart.Format(DateLayout), session.SessionTimeStart.Format(TimeLayout)
	}
	return "N/A", "N/A"
}
