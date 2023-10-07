package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

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
