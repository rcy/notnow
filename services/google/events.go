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

func (e *Event) AllDay() bool {
	return e.Start.DateTime == ""
}

func (e *Event) StartAt() time.Time {
	if e.Start.DateTime == "" {
		t, _ := time.Parse(time.DateOnly, e.Start.Date)
		return t
	}
	t, _ := time.Parse(time.RFC3339Nano, e.Start.DateTime)
	return t
}

func (e *Event) EndAt() time.Time {
	if e.End.DateTime == "" {
		t, _ := time.Parse(time.DateOnly, e.End.Date)
		return t
	}
	t, _ := time.Parse(time.RFC3339Nano, e.End.DateTime)
	return t
}

func (e *Event) Duration() time.Duration {
	return e.EndAt().Sub(e.StartAt())
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
