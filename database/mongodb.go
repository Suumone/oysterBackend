package database

import (
	"context"
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

func ConnectToMongoDB() *mongo.Client {
	uri := os.Getenv("DB_ADDRESS")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	MongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	err = MongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	return MongoClient
}

func CloseMongoDBConnection(mongoDB *mongo.Client) {
	err := mongoDB.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func SaveMentorInDB(user model.Users) (string, error) {
	collection := MongoDBClient.Database("Oyster").Collection("users")
	doc, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Printf("User(name: %s, insertedID: %s) inserted successfully\n", user.Username, doc.InsertedID)
	return doc.InsertedID.(primitive.ObjectID).Hex(), nil
}

func GetMentorsFromDB(params url.Values) []model.Users {
	collection := MongoDBClient.Database("Oyster").Collection("users")
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
	collection := MongoDBClient.Database("Oyster").Collection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": idToFind}
	var user model.Users
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("Document not found")
		} else {
			log.Printf("Failed to find document: %v\n", err)
		}
		return model.Users{}
	}
	return user
}

func UpdateMentorInDB(user model.Users, id string) error {
	idToFind, _ := primitive.ObjectIDFromHex(id)
	collection := MongoDBClient.Database("Oyster").Collection("users")
	filter := bson.M{"_id": idToFind}
	updateOp := bson.M{"$set": user}
	_, err := collection.UpdateOne(context.Background(), filter, updateOp)
	if err != nil {
		return err
	}

	log.Printf("User(id: %s) updated successfully!\n", id)
	return nil
}
