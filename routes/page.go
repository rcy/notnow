package routes

import (
	_ "embed"
	"html/template"
	"net/http"
	"time"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/layout"
	mw "yikes/middleware"
	"yikes/routes/tasks"
	"yikes/services/google"

	"google.golang.org/api/calendar/v3"
)

var (
	//go:embed page.gohtml
	pageContent  string
	pageTemplate = template.Must(template.New("mainpage").Parse(layout.Content + pageContent + tasks.Content))
)

func Page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := mw.UserFromContext(ctx)
	client, err := google.ClientForUser(ctx, user.ID)
	if err != nil {
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

	queries := yikes.New(db.Conn)
	tasks, err := queries.FindTasksByUserID(ctx, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = pageTemplate.Execute(w, struct {
		Events []*calendar.Event
		Tasks  []yikes.Task
	}{
		Events: events.Items,
		Tasks:  tasks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
