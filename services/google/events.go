package google

import (
	"context"
	"sort"
	"time"
	"yikes/db"
	"yikes/db/yikes"

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

func fetchFutureEvents(ctx context.Context, srv *calendar.Service) ([]Event, error) {
	gevents, err := srv.Events.
		List("primary").
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(time.Now().Add(28 * 24 * time.Hour).Format(time.RFC3339)).
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

func fetchPastEvents(ctx context.Context, srv *calendar.Service) ([]Event, error) {
	gevents, err := srv.Events.
		List("primary").
		TimeMax(time.Now().Format(time.RFC3339)).
		TimeMin(time.Now().Add(-28 * 24 * time.Hour).Format(time.RFC3339)).
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
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return fetchFutureEvents(ctx, srv)
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

func CreateTaskEvent(ctx context.Context, userID pgtype.UUID, task yikes.Task) (*calendar.Event, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	startAt, err := findNextAvailableTime(ctx, srv, task.Duration())
	if err != nil {
		return nil, err
	}

	event := calendar.Event{
		Summary:     "TASK " + task.Summary,
		Description: "--- yikes created event ---",
		Start: &calendar.EventDateTime{
			DateTime: startAt.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: startAt.Add(task.Duration()).Format(time.RFC3339),
		},
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"yikes": UUIDString(task.ID),
			},
		},
	}

	return srv.Events.Insert("primary", &event).Do()
}

func UUIDString(uuid pgtype.UUID) string {
	value, err := uuid.Value()
	if err != nil {
		return ""
	}
	str, _ := value.(string)
	return str
}

func StringUUID(str string) pgtype.UUID {
	uuid := pgtype.UUID{}
	_ = uuid.Scan(str)
	return uuid
}

func findNextAvailableTime(ctx context.Context, srv *calendar.Service, dur time.Duration) (*time.Time, error) {
	events, err := fetchFutureEvents(ctx, srv)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, event := range events {
		if now.Before(event.StartAt()) {
			// we are in a window of free time
			if event.StartAt().Sub(now) < dur {
				continue
			}
			return &now, nil
		}
		now = event.EndAt()
	}
	return &now, nil
}

func ReschedulePastTasks(ctx context.Context, userID pgtype.UUID) error {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return err
	}
	events, err := fetchPastEvents(ctx, srv)
	if err != nil {
		return err
	}

	queries := yikes.New(db.Conn)

	for _, ev := range events {
		if ev.ExtendedProperties != nil {
			str := ev.ExtendedProperties.Private["yikes"]
			if str == "" {
				continue
			}
			taskID := StringUUID(str)
			task, err := queries.UserTaskByID(ctx, yikes.UserTaskByIDParams{UserID: userID, ID: taskID})
			if err != nil {
				return err
			}

			startAt, err := findNextAvailableTime(ctx, srv, task.Duration())
			if err != nil {
				return err
			}
			ev.Start.DateTime = startAt.Format(time.RFC3339)
			ev.End.DateTime = startAt.Add(task.Duration()).Format(time.RFC3339)
			_, err = srv.Events.Update("primary", ev.Id, &ev.Event).Do()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
