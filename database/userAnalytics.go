package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"oysterProject/model"
)

func SaveAnalytics(userAnalytics *model.UserAnalytics) error {
	collection := GetCollection(UserAnalyticsCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	if userAnalytics.UserId == "" {
		_, err := collection.InsertOne(ctx, userAnalytics)
		log.Printf("Failed to insert userAnalytics(%s) error:%s\n", userAnalytics, err)
		return err
	} else {
		filter := bson.M{"userId": userAnalytics.UserId}
		updateOp := bson.M{"$set": userAnalytics}
		opts := options.Update().SetUpsert(true)
		_, err := collection.UpdateOne(ctx, filter, updateOp, opts)
		if err != nil {
			log.Printf("Failed to upadet userAnalytics (%s) error:%s\n", userAnalytics, err)
			return err
		}
	}
	return nil
}
