package schedulerJobs

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"oysterProject/database"
	"oysterProject/emailNotifications"
	"oysterProject/model"
	"time"
)

func sendUpcomingSessionNotification() {
	runJobWithTimeout(func(ctx context.Context) {
		sessionCollection := database.GetCollection(database.SessionCollectionName)
		currentTime := time.Now().UTC()
		filterPipeline := database.GetSessionsForNotificationPipeline(currentTime.Add(notificationTimeBeforeSession), currentTime.Add(2*notificationTimeBeforeSession))
		cursor, err := sessionCollection.Aggregate(ctx, filterPipeline)
		defer cursor.Close(ctx)
		if err != nil {
			log.Printf("SendUpcomingSessionNotification: Error executing search in db: %v", err)
			return
		}

		var sessions []model.SessionNotification
		if err = cursor.All(ctx, &sessions); err != nil {
			log.Printf("SendUpcomingSessionNotification: Failed to fetch sessions: %v", err)
			return
		}
		log.Printf("sendUpcomingSessionNotification count: %v\n", len(sessions))
		for _, session := range sessions {
			timeForNotification := session.SessionTimeStart.Sub(currentTime) - notificationTimeBeforeSession
			if timeForNotification <= 0 {
				continue
			}
			createRoutine(emailNotifications.SendNotificationBeforeSession, &session, timeForNotification)
		}
	})
}

func sendReviewEmails() {
	runJobWithTimeout(func(ctx context.Context) {
		sessionCollection := database.GetCollection(database.SessionCollectionName)
		filter := database.GetSessionsForReviewNotificationPipeline()
		cursor, err := sessionCollection.Aggregate(ctx, filter)
		defer cursor.Close(ctx)
		if err != nil {
			log.Printf("SendReviewEmails: Error executing search in db: %v", err)
			return
		}
		var sessions []model.SessionNotification
		if err = cursor.All(ctx, &sessions); err != nil {
			log.Printf("SendReviewEmails: Failed to fetch reviews: %v", err)
			return
		}
		log.Printf("sendReviewEmails count: %v\n", len(sessions))
		for _, session := range sessions {
			go emailNotifications.SendReviewEmails(&session)
		}
	})
}

func sendEmailForApprovedUsers() {
	runJobWithTimeout(func(ctx context.Context) {
		userCollection := database.GetCollection(database.UserCollectionName)
		filter := bson.M{
			"approvedEmailWasSent": false,
			"isApproved":           true,
		}
		cursor, err := userCollection.Find(ctx, filter)
		defer cursor.Close(ctx)
		if err != nil {
			log.Printf("sendEmailForApprovedUsers: Error executing search in db: %v", err)
			return
		}
		var users []model.User
		if err = cursor.All(ctx, &users); err != nil {
			log.Printf("sendEmailForApprovedUsers: Failed to fetch users: %v", err)
			return
		}
		log.Printf("sendEmailForApprovedUsers count: %v\n", len(users))
		for _, user := range users {
			go emailNotifications.SendApprovedEmail(&user)
		}
	})
}
