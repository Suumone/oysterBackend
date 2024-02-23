package schedulerJobs

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"oysterProject/database"
	"oysterProject/model"
	"time"
)

func statusCalculation() {
	runJobWithTimeout(func(ctx context.Context) {
		sessionCollection := database.GetCollection(database.SessionCollectionName)

		filterExpired := bson.M{
			"sessionTimeEnd": bson.M{"$lt": time.Now()},
			"sessionStatus":  bson.M{"$lt": model.Confirmed},
		}
		filterCompleted := bson.M{
			"sessionTimeStart": bson.M{"$lt": time.Now()},
			"sessionStatus":    model.Confirmed,
		}
		updateExpired := bson.M{"$set": bson.M{"sessionStatus": model.Expired}}
		updateCompleted := bson.M{"$set": bson.M{"sessionStatus": model.Completed}}

		runUpdateManyJob(ctx, sessionCollection, filterExpired, updateExpired, model.Expired.String())
		runUpdateManyJob(ctx, sessionCollection, filterCompleted, updateCompleted, model.Completed.String())
	})
}

func deleteExpired() {
	runJobWithTimeout(func(ctx context.Context) {
		collection := database.GetCollection(database.AuthSessionCollectionName)
		filter := bson.M{"expiry": bson.M{"$lt": time.Now().Unix()}}
		runDeleteManyJob(ctx, collection, filter)
	})
}

func runJobWithTimeout(jobFunc func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.TODO(), dbTimeout)
	defer cancel()
	jobFunc(ctx)
}

func runUpdateManyJob(ctx context.Context, collection *mongo.Collection, filter, update bson.M, statusType string) {
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Printf("UpdateMany job error: %v\n", err)
	}
	log.Printf("Matched %v documents and modified %v documents for %s status\n", result.MatchedCount, result.ModifiedCount, statusType)
}

func runDeleteManyJob(ctx context.Context, collection *mongo.Collection, filter bson.M) {
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("DeleteMany job error: %v\n", err)
	}
	log.Printf("Deleted count: %v\n", result.DeletedCount)
}
