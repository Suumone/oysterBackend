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

func CreateSession(session model.Session) (*model.SessionResponse, error) {
	mentorMenteeChan := make(chan []*model.UserImage)
	errChan := make(chan error)
	go func() {
		mentorMenteeImages, err := GetUserImages([]primitive.ObjectID{session.MentorId, session.MenteeId})
		if err != nil {
			errChan <- err
			return
		}
		mentorMenteeChan <- mentorMenteeImages
	}()

	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(SessionCollectionName)
	defer cancel()
	doc, err := collection.InsertOne(ctx, session)
	if err != nil {
		log.Printf("Error creating session: %v\n", err)
		return nil, err
	}
	session.SessionId = doc.InsertedID.(primitive.ObjectID)
	log.Printf("Session(menteeId: %s, mentorId: %s, sessionId:%s) created successfully\n", session.MenteeId, session.MentorId, doc.InsertedID)
	var mentorMenteeInfo []*model.UserImage
	select {
	case mentorMenteeFromChan := <-mentorMenteeChan:
		mentorMenteeInfo = mentorMenteeFromChan
	case errFromChan := <-errChan:
		if errFromChan != nil {
			return nil, utils.UserImageNotFound
		}
	}
	return createSessionResponse(mentorMenteeInfo, &session)
}

func GetSession(sessionId string) (*model.SessionResponse, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(SessionCollectionName)
	idToFind, _ := primitive.ObjectIDFromHex(sessionId)
	filter := bson.M{"_id": idToFind}
	var session model.Session
	err := collection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		handleFindError(err, sessionId, "session")
		return nil, err
	}

	mentorMenteeInfo, err := GetUserImages([]primitive.ObjectID{session.MentorId, session.MenteeId})
	if err != nil {
		return nil, err
	}
	return createSessionResponse(mentorMenteeInfo, &session)
}

func GetMentorMenteeIdsBySessionId(sessionId string) (*model.Session, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(SessionCollectionName)
	idToFind, _ := primitive.ObjectIDFromHex(sessionId)
	filter := bson.M{"_id": idToFind}
	var session model.Session
	err := collection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		handleFindError(err, sessionId, "session")
		return nil, err
	}
	return &session, nil
}

func createSessionResponse(mentorMenteeInfo []*model.UserImage, session *model.Session) (*model.SessionResponse, error) {
	var mentor, mentee *model.UserImage

	for _, userImage := range mentorMenteeInfo {
		if userImage.UserId == session.MentorId {
			mentor = userImage
			break
		}
	}

	for _, userImage := range mentorMenteeInfo {
		if userImage.UserId == session.MenteeId {
			mentee = userImage
			break
		}
	}
	model.SetStatusText(session)

	return &model.SessionResponse{
		SessionId:           session.SessionId,
		Mentor:              mentor,
		Mentee:              mentee,
		SessionTimeStart:    session.SessionTimeStart,
		SessionTimeEnd:      session.SessionTimeEnd,
		NewSessionTimeStart: session.NewSessionTimeStart,
		NewSessionTimeEnd:   session.NewSessionTimeEnd,
		RequestFromMentee:   session.RequestFromMentee,
		SessionStatus:       session.SessionStatus,
		Status:              session.Status,
		StatusForMentee:     session.StatusForMentee,
		StatusForMentor:     session.StatusForMentor,
		PaymentDetails:      session.PaymentDetails,
		MeetingLink:         session.MeetingLink,
		MenteeReview:        session.MenteeReview,
		MenteeRating:        session.MenteeRating,
	}, nil
}

func GetUserSessions(userId primitive.ObjectID, asMentor bool) ([]*model.SessionResponse, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(SessionCollectionName)
	filter := buildSessionFilter(userId, asMentor)
	cursor, err := collection.Find(ctx, filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		log.Printf("Failed to find sessions for mentee(%s): %v\n", userId.Hex(), err)
		return nil, err
	}
	defer cursor.Close(context.Background())
	return decodeSessions(cursor, true)
}

func GetUserUpcomingSessions(userId primitive.ObjectID, asMentor bool) ([]*model.SessionResponse, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(SessionCollectionName)
	filter := buildSessionFilter(userId, asMentor)
	filter["sessionStatus"] = bson.M{"$lte": model.Confirmed}

	cursor, err := collection.Find(ctx, filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		log.Printf("Failed to find sessions for mentee(%s): %v\n", userId.Hex(), err)
		return nil, err
	}
	defer cursor.Close(context.Background())
	return decodeSessions(cursor, false)
}

func buildSessionFilter(userId primitive.ObjectID, asMentor bool) bson.M {
	if asMentor {
		return bson.M{"mentorId": userId}
	}
	return bson.M{"menteeId": userId}
}

func decodeSessions(cursor *mongo.Cursor, withImage bool) ([]*model.SessionResponse, error) {
	var sessions []*model.SessionResponse
	for cursor.Next(context.Background()) {
		var session model.Session
		if err := cursor.Decode(&session); err != nil {
			log.Printf("Failed to decode session: %v\n", err)
			return nil, err
		}
		var mentorMenteeInfo []*model.UserImage
		if withImage {
			var err error
			mentorMenteeInfo, err = GetUserImages([]primitive.ObjectID{session.MentorId, session.MenteeId})
			if err != nil {
				return nil, err
			}
		}
		sessionResponse, err := createSessionResponse(mentorMenteeInfo, &session)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sessionResponse)
	}

	return sessions, nil
}

func RescheduleSession(session model.Session) (*model.SessionResponse, error) {
	filter := bson.M{"_id": session.SessionId}
	updateOp := bson.M{
		"$set": bson.M{
			"newSessionTimeStart": session.NewSessionTimeStart,
			"newSessionTimeEnd":   session.NewSessionTimeEnd,
			"sessionStatus":       session.SessionStatus,
		},
	}
	return updateSessionAndPrepareResponse(filter, updateOp)
}

func ConfirmSession(sessionId string) (*model.SessionResponse, error) {
	sessionIdObj, _ := primitive.ObjectIDFromHex(sessionId)
	filter := bson.M{"_id": sessionIdObj}

	session, err := GetSession(sessionId)
	if err != nil {
		return nil, err
	}
	var updateOp bson.M
	if session.SessionStatus == model.PendingByMentor {
		updateOp = bson.M{
			"$set": bson.M{
				"sessionTimeStart": session.SessionTimeStart,
				"sessionTimeEnd":   session.SessionTimeEnd,
				"sessionStatus":    model.Confirmed,
			},
		}
	} else {
		updateOp = bson.M{
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
	}

	return updateSessionAndPrepareResponse(filter, updateOp)
}

func CancelSession(sessionId, userId primitive.ObjectID) (*model.SessionResponse, error) {
	filter := bson.M{"_id": sessionId}

	user, err := GetUserByID(userId)
	if err != nil {
		log.Printf("CancelSession: Failed to find user(%s) err: %v\n", userId, err)
		return nil, err
	}
	var updateOp bson.M
	if user.AsMentor {
		updateOp = bson.M{"$set": bson.M{"sessionStatus": model.CanceledByMentor}}
	} else {
		updateOp = bson.M{"$set": bson.M{"sessionStatus": model.CanceledByMentee}}
	}

	return updateSessionAndPrepareResponse(filter, updateOp)
}

func updateSessionAndPrepareResponse(filter bson.M, updateOp bson.M) (*model.SessionResponse, error) {
	updatedSession, err := UpdateSession(filter, updateOp)
	if err != nil {
		return nil, err
	}
	mentorMenteeInfo, err := GetUserImages([]primitive.ObjectID{updatedSession.MentorId, updatedSession.MenteeId})
	if err != nil {
		return nil, err
	}
	return createSessionResponse(mentorMenteeInfo, updatedSession)
}

func UpdateSession(filter bson.M, updateOp bson.M) (*model.Session, error) {
	ctx, cancel := withTimeout(context.Background())
	defer cancel()
	collection := GetCollection(SessionCollectionName)
	var updatedSession model.Session
	err := collection.FindOneAndUpdate(
		ctx,
		filter,
		updateOp,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedSession)

	if err != nil {
		log.Printf("Failed to update session(%s) err: %v\n", filter["_id"], err)
		return nil, err
	}

	return &updatedSession, nil
}
