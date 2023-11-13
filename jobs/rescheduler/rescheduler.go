package rescheduler

import (
	"context"
	"log"
	"time"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/services/google"
)

func Loop() {
	ctx := context.Background()
	queries := yikes.New(db.Conn)

	for {
		users, err := queries.AllUsers(ctx)
		if err != nil {
			log.Printf("AllUsers error: %s", err)
			continue
		}

		for _, user := range users {
			log.Printf("rescheduling tasks for %s", user.Email)

			err := google.UpdateTaskDescriptions(ctx, user.ID)
			if err != nil {
				log.Printf("UpdateTaskDescriptions error: %s", err)
				continue
			}

			err = google.ReschedulePastTasks(ctx, user.ID)
			if err != nil {
				log.Printf("Reschedule error: %s", err)
				continue
			}
		}

		time.Sleep(5 * time.Minute)
	}
}
