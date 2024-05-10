package database

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	S3Client      *s3.S3
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

func ConnectToS3() {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(os.Getenv("DO_ACCESS_KEY"), os.Getenv("DO_SECRET_KEY"), ""),
		Endpoint:         aws.String(os.Getenv("DO_ENDPOINT")),
		Region:           aws.String(os.Getenv("DO_REGION")),
		S3ForcePathStyle: aws.Bool(false),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		log.Fatalf("Failed to disconnect from S3: %v", err)
	}
	S3Client = s3.New(newSession)
	log.Println("Connected to S3!")
}

func withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, dbTimeout)
}

func GetCollection(collectionName string) *mongo.Collection {
	return MongoDBOyster.Collection(collectionName)
}
