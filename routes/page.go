package routes

import (
	_ "embed"
	"html/template"
	"net/http"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/layout"
	mw "yikes/middleware"
	"yikes/routes/tasks"
	"yikes/services/google"
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

	events, err := google.Events(ctx, client)
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
		Events []google.Event
		Tasks  []yikes.Task
	}{
		Events: events,
		Tasks:  tasks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
