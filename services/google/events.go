package google

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
)

type Event struct {
	calendar.Event
}

func Events(ctx context.Context, client *http.Client) ([]Event, error) {
	srv, err := calendar.New(client)
	if err != nil {
		return nil, err
	}

	gevents, err := srv.Events.
		List("primary").
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	events := []Event{}
	for _, it := range gevents.Items {
		events = append(events, Event{*it})
	}

	return events, nil
}
