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

var (
	MongoDBClient *mongo.Client
	MongoDBOyster *mongo.Database
)

func ConnectToMongoDB() error {
	uri := os.Getenv("DB_ADDRESS")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var err error
	MongoDBClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	err = MongoDBClient.Ping(ctx, nil)
	if err != nil {
		return err
	}
	MongoDBOyster = MongoDBClient.Database("Oyster")
	log.Println("Connected to MongoDB!")
	return nil
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
