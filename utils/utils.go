package utils

import (
	"reflect"
	"runtime"
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

func TimePtr(t time.Time) *time.Time {
	return &t
}

func GetFunctionName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func convertUTCtoLocal(utcTime *time.Time, offset int) time.Time {
	loc := time.FixedZone("Local Timezone", offset)
	localTime := utcTime.In(loc)
	return localTime
}
