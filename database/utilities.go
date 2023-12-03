package database

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"strconv"
	"strings"
	"unicode"
)

// todo get from database
var fieldTypes = map[string]string{
	"language":                   "array",
	"countryDescription.country": "array",
	"mentorsTopics.topic":        "array",
	"areaOfExpertise.area":       "array",
	"experience":                 "number",
	"offset":                     "options",
	"limit":                      "options",
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

func handleFindError(err error, subject string) {
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		log.Printf("document(%s) not found", subject)
	default:
		log.Printf("Failed to find %s: %v", subject, err)
	}
}

func convertStringsToObjectIDs(stringSlice []string) ([]primitive.ObjectID, error) {
	var objectIDSlice []primitive.ObjectID
	for _, str := range stringSlice {
		objectID, err := primitive.ObjectIDFromHex(str)
		if err != nil {
			log.Printf("Failed to convert string(%s) to ObjectId: %v", str, err)
			return nil, err
		}
		objectIDSlice = append(objectIDSlice, objectID)
	}
	return objectIDSlice, nil
}
