package google

import (
	"context"
	"encoding/base32"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/calendar/v3"
)

const (
	taskPrefix = "."
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

type fetchEventsParam struct {
	Min time.Time
	Max time.Time
}

func fetchEvents(ctx context.Context, srv *calendar.Service, param fetchEventsParam) ([]Event, error) {
	gevents, err := srv.Events.
		List("primary").
		TimeMin(param.Min.Format(time.RFC3339)).
		TimeMax(param.Max.Format(time.RFC3339)).
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

func fetchFutureEvents(ctx context.Context, srv *calendar.Service) ([]Event, error) {
	return fetchEvents(ctx, srv, fetchEventsParam{
		Min: time.Now(),
		Max: time.Now().Add(28 * 24 * time.Hour),
	})
}

func fetchPastEvents(ctx context.Context, srv *calendar.Service) ([]Event, error) {
	return fetchEvents(ctx, srv, fetchEventsParam{
		Min: time.Now().Add(-28 * 24 * time.Hour),
		Max: time.Now(),
	})
}

func fetchAllEvents(ctx context.Context, srv *calendar.Service) ([]Event, error) {
	return fetchEvents(ctx, srv, fetchEventsParam{
		Min: time.Now().Add(-28 * 24 * time.Hour),
		Max: time.Now().Add(28 * 24 * time.Hour),
	})
}

func UserEvents(ctx context.Context, userID pgtype.UUID) ([]Event, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return fetchFutureEvents(ctx, srv)
}

func UserEvent(ctx context.Context, userID pgtype.UUID, eventID string) (*Event, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	gevent, err := srv.Events.Get("primary", eventID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	return &Event{*gevent}, nil
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

func CreateTaskEvent(ctx context.Context, userID pgtype.UUID, summary string, duration time.Duration) (*calendar.Event, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	startAt, err := findNextAvailableTime(ctx, srv, duration)
	if err != nil {
		return nil, err
	}

	// generate our own id so we can create a link back to it through this app
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	id := base32.
		NewEncoding("0123456789abcdefghijklmnopqrstuv").
		WithPadding(base32.NoPadding).
		EncodeToString(uuid[:])

	description := fmt.Sprintf("http://localhost:8080/event/%s", id)

	event := calendar.Event{
		Id:          id,
		Summary:     taskPrefix + summary,
		Description: description,
		Start: &calendar.EventDateTime{
			DateTime: startAt.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: startAt.Add(duration).Format(time.RFC3339),
		},
		// ExtendedProperties: &calendar.EventExtendedProperties{
		// 	Private: map[string]string{
		// 		"yikes": UUIDString(task.ID),
		// 	},
		// },
	}

	return srv.Events.Insert("primary", &event).Do()
}

func UpdateEventSummary(ctx context.Context, userID pgtype.UUID, eventID string, summary string) (*calendar.Event, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	event, err := UserEvent(ctx, userID, eventID)
	if err != nil {
		return nil, err
	}

	event.Summary = summary

	return srv.Events.Update("primary", eventID, &event.Event).Do()
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

func DeleteEvent(ctx context.Context, userID pgtype.UUID, eventID string) error {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return err
	}
	return srv.Events.Delete("primary", eventID).Do()
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

	//queries := yikes.New(db.Conn)

	for _, ev := range events {
		// task, err := queries.FindTaskByEventID(ctx, ev.Id)
		// if err != nil {
		// 	if err == pgx.ErrNoRows {
		// 		continue
		// 	}
		// 	return err
		// }

		if !strings.HasPrefix(ev.Summary, taskPrefix) {
			continue
		}

		dur := ev.Duration()

		startAt, err := findNextAvailableTime(ctx, srv, dur)
		if err != nil {
			return fmt.Errorf("ReschedulePastTasks: %w", err)
		}

		ev.Start.DateTime = startAt.Format(time.RFC3339)
		ev.End.DateTime = startAt.Add(dur).Format(time.RFC3339)
		log.Printf("%s startAt=%v ev.Duration()=%v %v %v", ev.Summary, startAt, ev.Duration(), ev.Start.DateTime, ev.End.DateTime)
		_, err = srv.Events.Update("primary", ev.Id, &ev.Event).Do()
		if err != nil {
			return fmt.Errorf("ReschedulePastTasks: %w", err)
		}
	}
	return nil
}

// Create tasks and add to database for all events that match taskPrefix
func MakeTasksForEvents(ctx context.Context, userID pgtype.UUID) error {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return err
	}
	events, err := fetchAllEvents(ctx, srv)
	if err != nil {
		return err
	}

	// tx, err := db.Conn.Begin(ctx)
	// if err != nil {
	// 	return err
	// }
	// defer tx.Rollback(ctx)

	// queries := yikes.New(tx)

	for _, ev := range events {
		log.Printf("ev: %s", ev.Summary)

		if !strings.HasPrefix(ev.Summary, taskPrefix) {
			continue
		}

		// _, err := queries.FindTaskByEventID(ctx, ev.Id)
		// if err == nil {
		// 	// we found a task
		// 	continue
		// } else {
		// 	if err != pgx.ErrNoRows {
		// 		// db error
		// 		return err
		// 	}
		// }

		// task, err := queries.CreateTask(ctx, yikes.CreateTaskParams{
		// 	UserID:  userID,
		// 	Summary: strings.TrimPrefix(ev.Summary, taskPrefix),
		// })
		// if err != nil {
		// 	return err
		// }

		// _, err = queries.CreateUserTaskEvent(ctx, yikes.CreateUserTaskEventParams{
		// 	UserID:  userID,
		// 	TaskID:  task.ID,
		// 	EventID: ev.Id,
		// })
		// if err != nil {
		// 	return err
		// }
	}

	//return tx.Commit(ctx)
	return nil
}
