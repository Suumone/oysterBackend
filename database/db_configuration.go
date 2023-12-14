package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

const dbTimeout = 10 * time.Second

var MongoDBClient *mongo.Client
var MongoDBOyster *mongo.Database

func ConnectToMongoDB() (*mongo.Client, error) {
	uri := os.Getenv("DB_ADDRESS")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	return mongoClient, nil
}

func CloseMongoDBConnection() {
	if err := MongoDBClient.Disconnect(context.Background()); err != nil {
		log.Fatalf("Failed to disconnect from MongoDB: %v", err)
	}
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, dbTimeout)
}

func GetCollection(collectionName string) *mongo.Collection {
	return MongoDBOyster.Collection(collectionName)
}
