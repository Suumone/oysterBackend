package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"oysterProject/model"
)

func CreateReview(review *model.Review) error {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(ReviewCollectionName)
	doc, err := collection.InsertOne(ctx, review)
	if err != nil {
		log.Printf("CreateReview: error creating session: %v\n", err)
		return err
	}
	review.ReviewId = doc.InsertedID.(primitive.ObjectID)
	return nil
}
