package schedulerJobs

import (
	"github.com/go-co-op/gocron"
	"log"
	"oysterProject/model"
	"time"
)

const (
	statusCalculationInterval     = 30 * time.Minute
	deleteExpiredSessionsInterval = 24 * time.Hour
	sendUpcomingSessionInterval   = 1 * time.Hour
	notificationTimeBeforeSession = 30 * time.Minute
	dbTimeout                     = 5 * time.Minute
)

var (
	notificationJobDelay = time.Duration(30-time.Now().Minute()%30) * time.Minute
)

func StartJobs() {
	startAsyncJob(statusCalculation, statusCalculationInterval, 0)
	startAsyncJob(deleteExpired, deleteExpiredSessionsInterval, 0)
	startAsyncJob(sendUpcomingSessionNotification, sendUpcomingSessionInterval, notificationJobDelay)
}

func startAsyncJob(jobFunc func(), interval, delay time.Duration) {
	j := gocron.NewScheduler(time.UTC)
	_, err := j.Every(interval).StartAt(time.Now().UTC().Add(delay)).Do(jobFunc)
	if err != nil {
		log.Fatalf("Error initializing job: %v\n", err)
		return
	}
	j.StartAsync()
}

func createRoutine(jobFunc func(session *model.SessionNotification), session *model.SessionNotification, delay time.Duration) {
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-timer.C:
			jobFunc(session)
		}
	}()
}
