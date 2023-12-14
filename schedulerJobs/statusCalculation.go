package schedulerJobs

import (
	"context"
	"errors"
	"github.com/go-co-op/gocron"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"oysterProject/database"
	"oysterProject/model"
	"time"
)

const INTERVAL = 30

func StartStatusCalculation() {
	j := gocron.NewScheduler(time.UTC)
	_, err := j.Every(INTERVAL).Minutes().Do(statusCalculation)
	if err != nil {
		log.Printf("Error initializing status calculation job: %v\n", err)
	}
	j.StartAsync()
}

func statusCalculation() {
	log.Println("statusCalculation")
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()
	sessionCollection := database.GetCollection("sessions")

	filter := bson.M{
		"sessionTimeEnd": bson.M{"$lt": time.Now()},
		"sessionStatus":  bson.M{"$lt": model.Confirmed},
	}
	update := bson.M{
		"$set": bson.M{"sessionStatus": model.Expired},
	}
	result, err := sessionCollection.UpdateMany(ctx, filter, update)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Printf("Status calculation job set expired status error: %v\n", err)
	}
	log.Printf("Matched %v documents and modified %v documents with expired status", result.MatchedCount, result.ModifiedCount)

	filter = bson.M{
		"sessionTimeStart": bson.M{"$lt": time.Now()},
		"sessionStatus":    model.Confirmed,
	}
	update = bson.M{
		"$set": bson.M{"sessionStatus": model.Completed},
	}
	result, err = sessionCollection.UpdateMany(ctx, filter, update)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		log.Printf("Status calculation job set completed status error: %v\n", err)
	}
	log.Printf("Matched %v documents and modified %v documents with completed status\n", result.MatchedCount, result.ModifiedCount)
}
