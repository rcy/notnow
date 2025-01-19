package rescheduler

import (
	"context"
	"log"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/services/google"
)

func RescheduleAll() error {
	ctx := context.Background()
	queries := yikes.New(db.Conn)

	users, err := queries.AllUsers(ctx)
	if err != nil {
		log.Printf("AllUsers error: %s", err)
		return err
	}

	for _, user := range users {
		log.Printf("rescheduling tasks for %s", user.Email)

		err := google.UpdateTaskDescriptions(ctx, user.ID)
		if err != nil {
			log.Printf("UpdateTaskDescriptions error: %s", err)
			return err
		}

		err = google.ReschedulePastTasks(ctx, user.ID)
		if err != nil {
			log.Printf("Reschedule error: %s", err)
			return err
		}
	}

	return nil
}
