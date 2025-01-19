package google

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

type EventModel struct {
	calendar.Event
}

func (e *EventModel) AllDay() bool {
	return e.Start.DateTime == ""
}

func (e *EventModel) StartDate() string {
	if e.Start.Date != "" {
		return e.Start.Date
	}
	return e.StartAt().Format(time.DateOnly)
}

func (e *EventModel) StartTime() string {
	if e.Start.Date != "" {
		return ""
	}
	return e.StartAt().Format("15:04")
}

func (e *EventModel) EndTime() string {
	if e.End.Date != "" {
		return ""
	}
	return e.EndAt().Format("15:04")
}

func (e *EventModel) StartAt() time.Time {
	if e.Start.DateTime == "" {
		t, _ := time.Parse(time.DateOnly, e.Start.Date)
		return t
	}
	t, _ := time.Parse(time.RFC3339Nano, e.Start.DateTime)
	return t
}

func (e *EventModel) EndAt() time.Time {
	if e.End.DateTime == "" {
		t, _ := time.Parse(time.DateOnly, e.End.Date)
		return t
	}
	t, _ := time.Parse(time.RFC3339Nano, e.End.DateTime)
	return t
}

func (e *EventModel) Duration() time.Duration {
	return e.EndAt().Sub(e.StartAt())
}

func (e *EventModel) IsTask() bool {
	return strings.HasPrefix(e.Summary, taskPrefix)
}

var contextRegexp = regexp.MustCompile(regexp.QuoteMeta(withContextPrefix) + "(\\w+)")

func (e *EventModel) Contexts() []string {
	contexts := []string{}
	for _, matches := range contextRegexp.FindAllStringSubmatch(e.Summary, -1) {
		contexts = append(contexts, matches[1])
	}
	return contexts
}

func (e *EventModel) IsContainer() bool {
	return strings.HasPrefix(e.Summary, containerPrefix)
}

var containerRegexp = regexp.MustCompile(regexp.QuoteMeta(containerPrefix) + "(\\w+)")

func (e *EventModel) ContainerContexts() []string {
	contexts := []string{}
	for _, matches := range containerRegexp.FindAllStringSubmatch(e.Summary, -1) {
		contexts = append(contexts, matches[1])
	}
	return contexts
}

func (e *EventModel) IsContainerFor(context string) bool {
	for _, c := range e.ContainerContexts() {
		if c == context {
			return true
		}
	}
	return false
}

// Return the url to the event in this application
func (e *EventModel) AppURL() string {
	return fmt.Sprintf("%s/events/%s", os.Getenv("ROOT_URL"), e.Id)
}
