package google

import (
	"testing"

	"golang.org/x/exp/slices"
	"google.golang.org/api/calendar/v3"
)

func TestEventContext(t *testing.T) {
	for _, tc := range []struct {
		summary string
		want    []string
	}{
		{"fold socks", nil},
		{"buy batteries +town", []string{"town"}},
		{"chop wood +chore +outside", []string{"chore", "outside"}},
		{"+barecontext", []string{"barecontext"}},
		{"starts with number +9abc", []string{"9abc"}},
		{"ends with number +abc9", []string{"abc9"}},
	} {
		t.Run(tc.summary, func(t *testing.T) {
			e := Event{
				Event: calendar.Event{
					Summary: tc.summary,
				},
			}
			got := e.Contexts()
			if !slices.Equal(got, tc.want) {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestEventContainerContexts(t *testing.T) {
	for _, tc := range []struct {
		summary string
		want    []string
	}{
		{"meeting with boss", nil},
		{"=work", []string{"work"}},
		{"=read", []string{"read"}},
		{"=read books or websites", []string{"read"}},
		{"=read =watch just consume things", []string{"read", "watch"}},
	} {
		t.Run(tc.summary, func(t *testing.T) {
			e := Event{
				Event: calendar.Event{
					Summary: tc.summary,
				},
			}
			if got := e.ContainerContexts(); !slices.Equal(got, tc.want) {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestEventIsContainer(t *testing.T) {
	for _, tc := range []struct {
		summary string
		want    bool
	}{
		{"meeting with boss", false},
		{"=work", true},
		{"=read", true},
		{"=read books or websites", true},
		{"=read =watch just consume things", true},
	} {
		t.Run(tc.summary, func(t *testing.T) {
			e := Event{
				Event: calendar.Event{
					Summary: tc.summary,
				},
			}
			if got := e.IsContainer(); got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestEventIsContainerFor(t *testing.T) {
	for _, tc := range []struct {
		summary string
		context string
		want    bool
	}{
		{"meeting with boss", "work", false},
		{"=work", "work", true},
		{"=work", "play", false},
		{"=read", "read", true},
		{"=read books or websites", "read", true},
		{"=read =watch just consume things", "read", true},
		{"=read just consume things =watch", "watch", true},
		{"=read =watch just consume things", "work", false},
		{"=foo", "", false},
		{"=foo", "foo", true},
		//{"==foo", "foo", false},
	} {
		t.Run(tc.summary, func(t *testing.T) {
			e := Event{
				Event: calendar.Event{
					Summary: tc.summary,
				},
			}
			if got := e.IsContainerFor(tc.context); got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
