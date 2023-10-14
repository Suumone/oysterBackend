package database

import (
	"strconv"
	"strings"
	"unicode"
)

// todo get from database
var fieldTypes = map[string]string{
	"language":                   "array",
	"countryDescription.country": "array",
	"mentorsTopics.topic":        "array",
	"experience":                 "number",
}

func convertStringToNumber(s string) float32 {
	cleanString := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) || r == '.' || r == '-' {
			return r
		}
		return -1
	}, s)

	f, err := strconv.ParseFloat(cleanString, 32)
	if err != nil {
		return 0
	}

	return float32(f)
}
