package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"oysterProject/model"
	"time"
)

func SaveSession(s *model.AuthSession) (string, error) {
	collection := GetCollection(AuthSessionCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	result, err := collection.InsertOne(ctx, s)
	if err != nil {
		log.Printf("Error saving auth session in db: %v\n", err)
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

func UpdateSession(sessionId primitive.ObjectID, expiryTime time.Time) error {
	collection := GetCollection(AuthSessionCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	updateOp := bson.M{"$set": bson.M{"expiry": expiryTime.Unix()}}
	result, err := collection.UpdateByID(ctx, sessionId, updateOp)
	if err != nil || result.ModifiedCount == 0 {
		log.Printf("Error updeting auth session in db: %v\n", err)
		return err
	}
	return nil
}

func FindSession(sessionId string) (*model.AuthSession, bool) {
	collection := GetCollection(AuthSessionCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	sessionIdObj, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		log.Printf("Failed to convert string identifier to object(%s): %v\n", sessionId, err)
		return nil, false
	}
	filter := bson.M{"_id": sessionIdObj}
	var s model.AuthSession
	err = collection.FindOne(ctx, filter).Decode(&s)
	if err != nil {
		log.Printf("Auth session was not found(%s): %v\n", sessionId, err)
		return nil, false
	}
	if time.Now().Unix() > s.Expiry {
		log.Printf("Auth session(%s) expired\n", sessionId)
		return nil, false
	}
	return &s, true
}

func DeleteSession(s *model.AuthSession) error {
	collection := GetCollection(AuthSessionCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	filter := bson.M{"_id": s.SessionId}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil || result.DeletedCount == 0 {
		log.Printf("Error deleting session(%s): %v\n", s.SessionId, err)
		return err
	}
	return nil
}
