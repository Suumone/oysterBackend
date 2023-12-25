package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"oysterProject/model"
	"oysterProject/utils"
	"time"
)

func CreateReviewAndUpdateSession(sessionReview *model.SessionReview) (*model.SessionResponse, error) {
	session, err := UpdateSessionReviews(sessionReview)
	if err != nil {
		return nil, nil
	}
	review := &model.Review{
		MenteeId: session.Mentee.UserId,
		MentorId: session.Mentor.UserId,
		Review:   sessionReview.PublicReview,
		Rating:   sessionReview.PublicRating,
		Date:     utils.TimePtr(time.Now()),
	}
	err = CreateReview(review)
	return &session, err
}

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
