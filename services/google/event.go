package google

import (
	"strings"
	"time"

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

func (e *Event) IsTask() bool {
	return strings.HasPrefix(e.Summary, taskPrefix)
}

func (e *Event) IsContext() bool {
	return strings.HasPrefix(e.Summary, contextPrefix)
}
