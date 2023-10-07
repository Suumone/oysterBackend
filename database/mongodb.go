package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/url"
	"os"
	"oysterProject/model"
	"time"
)

var MongoDBClient *mongo.Client
var MongoDBOyster *mongo.Database

func ConnectToMongoDB() *mongo.Client {
	uri := os.Getenv("DB_ADDRESS")
	Context, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	MongoClient, err := mongo.Connect(Context, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	err = MongoClient.Ping(Context, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	return MongoClient
}

func CloseMongoDBConnection() {
	if err := MongoDBClient.Disconnect(context.Background()); err != nil {
		log.Fatalf("Failed to disconnect from MongoDB: %v", err)
	}
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
		filter[key] = bson.M{"$all": values}
	}
	log.Printf("MongoDB filter:%s\n", filter)
	return filter
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

	pipeline := bson.A{
		bson.D{{"$match", bson.D{{"_id", idToFind}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "reviews"},
					{"localField", "_id"},
					{"foreignField", "user"},
					{"as", "reviews"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$reviews"}}}},
		bson.D{
			{"$lookup",
				bson.D{
					{"from", "users"},
					{"localField", "reviews.reviewer"},
					{"foreignField", "_id"},
					{"as", "reviewerInfo"},
				},
			},
		},
		bson.D{{"$unwind", bson.D{{"path", "$reviewerInfo"}}}},
		bson.D{
			{"$project",
				bson.D{
					{"id", "$user"},
					{"reviews",
						bson.D{
							{"review", "$reviews.review"},
							{"rating", "$reviews.rating"},
							{"date", "$reviews.date"},
							{"reviewer",
								bson.D{
									{"id", "$reviewerInfo._id"},
									{"name", "$reviewerInfo.name"},
									{"jobTitle", "$reviewerInfo.jobTitle"},
									{"profileImage", "$reviewerInfo.profileImage"},
								},
							},
						},
					},
				},
			},
		},
		bson.D{
			{"$group",
				bson.D{
					{"_id", "$_id"},
					{"reviews", bson.D{{"$push", "$reviews"}}},
				},
			},
		},
	}
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
