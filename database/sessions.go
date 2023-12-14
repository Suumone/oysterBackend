package database

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"oysterProject/model"
	"oysterProject/utils"
)

func CreateSession(session model.Session) (primitive.ObjectID, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("sessions")
	defer cancel()
	doc, err := collection.InsertOne(ctx, session)
	if err != nil {
		log.Printf("Error creating session: %v\n", err)
		return primitive.ObjectID{}, err
	}
	log.Printf("Session(menteeId: %s, mentorId: %s, sessionId:%s) created successfully\n", session.MenteeId, session.MentorId, doc.InsertedID)
	return doc.InsertedID.(primitive.ObjectID), nil
}

func GetSession(sessionId string) (model.Session, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("sessions")
	idToFind, _ := primitive.ObjectIDFromHex(sessionId)
	filter := bson.M{"_id": idToFind}
	var session model.Session
	err := collection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		handleFindError(err, sessionId, "session")
		return model.Session{}, err
	}
	return session, nil
}

func GetUserSessions(userId primitive.ObjectID, asMentor bool) ([]model.Session, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection("sessions")
	filter := buildSessionFilter(userId, asMentor)
	cursor, err := collection.Find(ctx, filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		log.Printf("Failed to find sessions for mentee(%s): %v\n", userId.Hex(), err)
		return nil, err
	}
	defer cursor.Close(context.Background())
	return decodeSessions(cursor)
}

func buildSessionFilter(userId primitive.ObjectID, asMentor bool) bson.M {
	if asMentor {
		return bson.M{"mentorId": userId}
	}
	return bson.M{"menteeId": userId}
}

func decodeSessions(cursor *mongo.Cursor) ([]model.Session, error) {
	var sessions []model.Session
	for cursor.Next(context.Background()) {
		var session model.Session
		if err := cursor.Decode(&session); err != nil {
			log.Printf("Failed to decode session: %v\n", err)
			return nil, err
		}

		utils.GetStatusText(&session)
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func RescheduleSession(session model.Session) (model.Session, error) {
	collection := GetCollection("sessions")
	filter := bson.M{"_id": session.SessionId}
	updateOp := bson.M{
		"$set": bson.M{
			"newSessionTimeStart": session.NewSessionTimeStart,
			"newSessionTimeEnd":   session.NewSessionTimeEnd,
			"sessionStatus":       session.SessionStatus,
		},
	}

	return updateSession(collection, filter, updateOp)
}

func ConfirmSession(sessionId string) (model.Session, error) {
	collection := GetCollection("sessions")
	sessionIdObj, _ := primitive.ObjectIDFromHex(sessionId)
	filter := bson.M{"_id": sessionIdObj}

	session, err := GetSession(sessionId)
	if err != nil {
		return model.Session{}, err
	}

	updateOp := bson.M{
		"$set": bson.M{
			"sessionTimeStart": session.NewSessionTimeStart,
			"sessionTimeEnd":   session.NewSessionTimeEnd,
			"sessionStatus":    model.Confirmed,
		},
		"$unset": bson.M{
			"newSessionTimeStart": "",
			"newSessionTimeEnd":   "",
		},
	}

	return updateSession(collection, filter, updateOp)
}

func CancelSession(sessionId, userId string) (model.Session, error) {
	collection := GetCollection("sessions")
	sessionIdObj, _ := primitive.ObjectIDFromHex(sessionId)
	filter := bson.M{"_id": sessionIdObj}

	user := GetUserByID(userId)
	var updateOp bson.M
	if user.AsMentor {
		updateOp = bson.M{"$set": bson.M{"sessionStatus": model.CanceledByMentor}}
	} else {
		updateOp = bson.M{"$set": bson.M{"sessionStatus": model.CanceledByMentee}}
	}

	return updateSession(collection, filter, updateOp)
}

func updateSession(collection *mongo.Collection, filter bson.M, updateOp bson.M) (model.Session, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()

	var updatedSession model.Session
	err := collection.FindOneAndUpdate(
		ctx,
		filter,
		updateOp,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedSession)

	if err != nil {
		log.Printf("Failed to update session(%s) err: %v\n", filter["_id"], err)
		return model.Session{}, err
	}

	return updatedSession, nil
}