package google

import (
	"context"
	"encoding/base32"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/calendar/v3"
)

const (
	taskPrefix        = "."
	containerPrefix   = "="
	withContextPrefix = "+"
)

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

func eventURL(id string) string {
	return fmt.Sprintf("%s/event/%s", os.Getenv("ROOT_URL"), id)
}

func CreateTaskEvent(ctx context.Context, userID pgtype.UUID, summary string, duration time.Duration) (*calendar.Event, error) {
	srv, err := ServiceForUser(ctx, userID)
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

	dummyStartAt := time.Now().Add(-24 * time.Hour)
	event := calendar.Event{
		Id:          id,
		Summary:     taskPrefix + summary,
		Description: eventURL(id),
		// need to provide a start and end time so duration can be computed
		Start: &calendar.EventDateTime{DateTime: dummyStartAt.Format(time.RFC3339)},
		End:   &calendar.EventDateTime{DateTime: dummyStartAt.Add(duration).Format(time.RFC3339)},
	}

	startAt, err := findNextAvailableTime(ctx, srv, &Event{event})
	if err != nil {
		return nil, err
	}

	event.Start = &calendar.EventDateTime{DateTime: startAt.Format(time.RFC3339)}
	event.End = &calendar.EventDateTime{DateTime: startAt.Add(duration).Format(time.RFC3339)}

	return srv.Events.Insert("primary", &event).Do()
}

func UpdateEventSummary(ctx context.Context, userID pgtype.UUID, eventID string, summary string) (*calendar.Event, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	event := &calendar.Event{
		Summary: summary,
	}

	return srv.Events.Patch("primary", eventID, event).Do()
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

func findNextAvailableTime(ctx context.Context, srv *calendar.Service, ev *Event) (*time.Time, error) {
	events, err := fetchFutureEvents(ctx, srv)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// start at the time of the first matching context
	if contexts := ev.Contexts(); len(contexts) > 0 {
		// TODO consider all contexts, not just first one
		context := contexts[0]
		containerEvent := findContextContainer(events, context)
		now = containerEvent.StartAt()

		// trim out the older events
		newEvents := []Event{}
		for _, ev := range events {
			if !ev.StartAt().Before(containerEvent.StartAt()) {
				newEvents = append(newEvents, ev)
			}
		}
		events = newEvents
	}

	for _, event := range events {
		if now.Before(event.StartAt()) {
			// we are in a window of free time
			if event.StartAt().Sub(now) < ev.Duration() {
				continue
			}
			return &now, nil
		}
		now = event.EndAt()
	}
	return &now, nil
}

// Return first event that is a container for context
func findContextContainer(events []Event, context string) *Event {
	for _, ev := range events {
		if ev.IsContainerFor(context) {
			return &ev
		}
	}
	return nil
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

	for _, ev := range events {
		if !strings.HasPrefix(ev.Summary, taskPrefix) {
			continue
		}

		dur := ev.Duration()

		startAt, err := findNextAvailableTime(ctx, srv, &ev)
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

func UpdateTaskDescriptions(ctx context.Context, userID pgtype.UUID) error {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return err
	}
	events, err := fetchAllEvents(ctx, srv)
	if err != nil {
		return err
	}

	for _, ev := range events {
		log.Printf("ev: %s", ev.Summary)

		if !strings.HasPrefix(ev.Summary, taskPrefix) {
			continue
		}

		err := updateTaskDescription(srv, &ev)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update event to include link in the description to this event if it does not already exist
func updateTaskDescription(srv *calendar.Service, ev *Event) error {
	url := eventURL(ev.Id)

	match, err := regexp.MatchString(regexp.QuoteMeta(url), ev.Description)
	if err != nil {
		return err
	}
	if match {
		return nil
	}
	ev.Description = fmt.Sprintf("%s\n%s", url, ev.Description)
	gev, err := srv.Events.Update("primary", ev.Id, &ev.Event).Do()
	if err != nil {
		return err
	}
	ev.Event = *gev
	return err
}
