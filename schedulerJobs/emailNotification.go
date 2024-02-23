package schedulerJobs

import (
	"context"
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
		filterPipeline := database.GetSessionsForNotification(currentTime.Add(notificationTimeBeforeSession), currentTime.Add(2*notificationTimeBeforeSession))
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

		for _, session := range sessions {
			timeForNotification := session.SessionTimeStart.Sub(currentTime) - notificationTimeBeforeSession
			if timeForNotification <= 0 {
				continue
			}
			createRoutine(emailNotifications.SendNotificationBeforeSession, &session, timeForNotification)
		}
	})
}
