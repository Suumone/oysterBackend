package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"oysterProject/model"
	"time"
)

func SaveAuthSession(s *model.AuthSession) (string, error) {
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

func UpdateAuthSession(sessionId primitive.ObjectID, expiryTime time.Time) (*model.AuthSession, error) {
	collection := GetCollection(AuthSessionCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	filter := bson.M{"_id": sessionId}
	updateOp := bson.M{"$set": bson.M{"expiry": expiryTime.Unix()}}
	var result model.AuthSession
	err := collection.FindOneAndUpdate(ctx, filter, updateOp, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&result)
	if err != nil {
		log.Printf("Error updeting auth session in db: %v\n", err)
		return nil, err
	}
	return &result, nil
}

func FindAuthSession(sessionId primitive.ObjectID) (*model.AuthSession, bool) {
	collection := GetCollection(AuthSessionCollectionName)
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	filter := bson.M{"_id": sessionId}
	var s model.AuthSession
	err := collection.FindOne(ctx, filter).Decode(&s)
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

func DeleteAuthSession(s *model.AuthSession) error {
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
