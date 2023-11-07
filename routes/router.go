package routes

import (
	_ "embed"
	"html/template"
	"net/http"
	"time"
	mw "yikes/middleware"
	"yikes/routes/auth"

	"github.com/go-chi/chi/v5"
	"google.golang.org/api/calendar/v3"
)

func Router(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(mw.User)
		r.Use(mw.Client)
		r.Get("/", Page)
	})

	r.Route("/auth", auth.Router)
}

var (
	//go:embed page.gohtml
	pageContent  string
	pageTemplate = template.Must(template.New("page").Parse(pageContent))
)

func Page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client, ok := mw.ClientFromContext(ctx)
	if !ok {
		http.Error(w, "couldn't get client", http.StatusInternalServerError)
		return
	}

	srv, err := calendar.New(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	events, err := srv.Events.
		List("primary").
		TimeMin(time.Now().Format(time.RFC3339)).
		TimeMax(time.Now().Add(365 * 24 * time.Hour).Format(time.RFC3339)).
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx).
		Do()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = pageTemplate.Execute(w, struct {
		Events []*calendar.Event
	}{
		Events: events.Items,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
