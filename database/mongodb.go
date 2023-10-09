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
	"time"
	"unicode"
)

var MongoDBClient *mongo.Client
var MongoDBOyster *mongo.Database

const dbTimeout = 5 * time.Second

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, dbTimeout)
}

func GetCollection(collectionName string) *mongo.Collection {
	return MongoDBOyster.Collection(collectionName)
}

func SaveMentor(user model.User) (primitive.ObjectID, error) {
	collection := GetCollection("users")
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	doc, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Error inserting user: %v\n", err)
		return primitive.ObjectID{}, err
	}
	log.Printf("User(name: %s, insertedID: %s) inserted successfully\n", user.Username, doc.InsertedID)
	return doc.InsertedID.(primitive.ObjectID), nil
}

func GetMentors(params url.Values) []model.User {
	collection := MongoDBOyster.Collection("users")
	filter := getFilterQueryFromUrlParams(params)
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to find documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.Background())

	var users []model.User
	for cursor.Next(context.Background()) {
		var user model.User
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

func GetUserByID(id string) model.User {
	collection := GetCollection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": idToFind}
	var user model.User
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Println("Document not found")
		} else {
			log.Printf("Failed to find document: %v\n", err)
		}
		return model.User{}
	}
	return user
}

func GetMentorReviewsByID(id string) model.UserWithReviews {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	usersColl := GetCollection("users")
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

func UpdateMentor(user model.User, id string) error {
	user.IsNewUser = false
	idToFind, _ := primitive.ObjectIDFromHex(id)
	collection := GetCollection("users")
	filter := bson.M{"_id": idToFind}
	updateOp := bson.M{"$set": user}
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	_, err := collection.UpdateOne(ctx, filter, updateOp)
	if err != nil {
		return err
	}

	log.Printf("User(id: %s) updated successfully!\n", id)
	return nil
}

func GetListOfFilterFields() ([]map[string]interface{}, error) {
	var fields []map[string]interface{}
	filterColl := GetCollection("fieldInfo")
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	cursor, err := filterColl.Find(ctx, bson.D{})
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

func GetReviewsForFrontPage() []model.ReviewsForFrontPage {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	reviewColl := GetCollection("reviews")
	pipeline := GetFrontPageReviewsPipeline()
	cursor, err := reviewColl.Aggregate(ctx, pipeline)
	if err != nil {
		return []model.ReviewsForFrontPage{}
	}
	defer cursor.Close(ctx)

	var result []model.ReviewsForFrontPage
	for cursor.Next(ctx) {
		var review model.ReviewsForFrontPage
		err := cursor.Decode(&review)
		if err != nil {
			return []model.ReviewsForFrontPage{}
		}
		result = append(result, review)
	}
	return result
}

func GetUserByEmail(user model.User) (model.User, error) {
	collection := GetCollection("users")
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	filter := bson.M{"email": user.Email}
	err := collection.FindOne(ctx, filter).Decode(&user)
	return user, err
}
