package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/url"
	"oysterProject/model"
	"strconv"
	"strings"
	"unicode"
)

var MongoDBClient *mongo.Client
var MongoDBOyster *mongo.Database

func GetCollection(collectionName string) *mongo.Collection {
	return MongoDBOyster.Collection(collectionName)
}

func SaveMentorInDB(user model.Users) (string, error) {
	collection := MongoDBOyster.Collection("users")
	doc, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Printf("User(name: %s, insertedID: %s) inserted successfully\n", user.Username, doc.InsertedID)
	return doc.InsertedID.(primitive.ObjectID).Hex(), nil
}

func GetMentorsFromDB(params url.Values) []model.Users {
	collection := MongoDBOyster.Collection("users")
	filter := getFilterQueryFromUrlParams(params)
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to find documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.Background())

	var users []model.Users
	for cursor.Next(context.Background()) {
		var user model.Users
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode document: %v", err)
		}
		users = append(users, user)
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
	}
	return users
}

func getFilterQueryFromUrlParams(params url.Values) bson.M {
	filter := bson.M{}
	filter["mentor"] = true
	for key, values := range params {
		if key == "experience" {
			filter[key] = bson.M{"$gt": convertStringToNumber(values[0])}
		} else {
			filter[key] = bson.M{"$all": values}
		}
	}
	log.Printf("MongoDB filter:%s\n", filter)
	return filter
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

func GetMentorByIDFromDB(id string) model.Users {
	collection := MongoDBOyster.Collection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": idToFind}
	var user model.Users
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Println("Document not found")
		} else {
			log.Printf("Failed to find document: %v\n", err)
		}
		return model.Users{}
	}
	return user
}

func GetMentorReviewsByIDFromDB(id string) model.UserWithReviews {
	ctx := context.Background()
	usersColl := MongoDBOyster.Collection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)

	pipeline := GetMentorListPipeline(idToFind)
	cursor, err := usersColl.Aggregate(ctx, pipeline)
	if err != nil {
		return model.UserWithReviews{}
	}
	defer cursor.Close(ctx)
	var user model.UserWithReviews
	for cursor.Next(context.Background()) {
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode document: %v", err)
		}
	}

	return user
}

func UpdateMentorInDB(user model.Users, id string) error {
	idToFind, _ := primitive.ObjectIDFromHex(id)
	collection := MongoDBOyster.Collection("users")
	filter := bson.M{"_id": idToFind}
	updateOp := bson.M{"$set": user}
	_, err := collection.UpdateOne(context.Background(), filter, updateOp)
	if err != nil {
		return err
	}

	log.Printf("User(id: %s) updated successfully!\n", id)
	return nil
}

func GetListOfFilterFields() ([]map[string]interface{}, error) {
	var fields []map[string]interface{}
	filterColl := GetCollection("fieldInfo")
	cursor, err := filterColl.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}

	var metas []map[string]interface{}
	if err = cursor.All(context.TODO(), &metas); err != nil {
		return nil, err
	}

	for _, meta := range metas {
		fieldData, err := extractFieldDataFromMeta(meta)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fieldData)
	}

	return fields, nil
}

func extractFieldDataFromMeta(meta map[string]interface{}) (map[string]interface{}, error) {
	fieldName := meta["fieldName"].(string)
	fieldType := meta["type"].(string)
	fieldStorage := meta["fieldStorage"].(string)

	valuesFromDb, ok := meta["values"].(primitive.A)
	if !ok {
		valuesFromDb = primitive.A{}
	}

	fieldData := map[string]interface{}{
		"fieldName":    fieldName,
		"type":         fieldType,
		"fieldStorage": fieldStorage,
		"values":       valuesFromDb,
	}

	if fieldType == "dropdown" && len(valuesFromDb) == 0 {
		usersColl := GetCollection("users")
		values, err := usersColl.Distinct(context.TODO(), fieldStorage, bson.D{})
		if err != nil {
			return nil, err
		}
		fieldData["values"] = values
	}

	return fieldData, nil
}
