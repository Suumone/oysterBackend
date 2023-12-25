package utils

import (
	"net/url"
	"oysterProject/model"
	"reflect"
	"strings"
	"time"
)

const ImageLimitSizeMB = 1024 * 1024 * 5 //5 MB

func IsEmptyStruct(input interface{}) bool {
	zeroValue := reflect.New(reflect.TypeOf(input)).Elem().Interface()
	return reflect.DeepEqual(input, zeroValue)
}

func NormalizeSocialLinks(user *model.User) {
	user.LinkedInLink = makeURL(user.LinkedInLink, "linkedin.com/")
	user.InstagramLink = makeURL(user.InstagramLink, "instagram.com/")
	user.FacebookLink = makeURL(user.FacebookLink, "facebook.com/")
	user.CalendlyLink = makeURL(user.CalendlyLink, "calendly.com/")
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
