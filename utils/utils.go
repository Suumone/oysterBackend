package utils

import (
	"net/url"
	"oysterProject/model"
	"reflect"
	"strings"
)

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
