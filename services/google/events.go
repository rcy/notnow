package google

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/calendar/v3"
)

type Event struct {
	calendar.Event
}

func (e *Event) AllDay() bool {
	return e.Start.DateTime == ""
}

func (e *Event) StartDate() string {
	if e.Start.Date != "" {
		return e.Start.Date
	}
	return e.StartAt().Format(time.DateOnly)
}

func (e *Event) StartTime() string {
	if e.Start.Date != "" {
		return ""
	}
	return e.StartAt().Format("15:04")
}

func (e *Event) EndTime() string {
	if e.End.Date != "" {
		return ""
	}
	return e.EndAt().Format("15:04")
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

func clientEvents(ctx context.Context, client *http.Client) ([]Event, error) {
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

func UserEvents(ctx context.Context, userID pgtype.UUID) ([]Event, error) {
	client, err := ClientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return clientEvents(ctx, client)
}

type TimeGrouping struct {
	Events map[string][]Event
	Keys   []string
}

// Return events for user grouped by day.
func UserEventsGroupedByDay(ctx context.Context, userID pgtype.UUID) (*TimeGrouping, error) {
	events, err := UserEvents(ctx, userID)
	if err != nil {
		return nil, err
	}

	tg := TimeGrouping{
		Events: map[string][]Event{},
		Keys:   []string{},
	}

	for _, e := range events {
		day := e.StartDate()
		tg.Events[day] = append(tg.Events[day], e)
	}

	for key := range tg.Events {
		tg.Keys = append(tg.Keys, key)
	}

	sort.Slice(tg.Keys, func(i, j int) bool {
		return tg.Keys[i] < tg.Keys[j]
	})

	return &tg, nil
}
