package routes

import (
	_ "embed"
	"html/template"
	"net/http"
	"yikes/layout"
	mw "yikes/middleware"
	"yikes/routes/events"
	"yikes/routes/tasks"
	"yikes/services/google"
)

var (
	//go:embed page.gohtml
	pageContent  string
	pageTemplate = template.Must(template.New("mainpage").Funcs(events.Funcs).Parse(layout.Content + pageContent + events.Content + tasks.Content))
)

func Page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := mw.UserFromContext(ctx)

	groupedEvents, err := google.UserEventsGroupedByDay(ctx, user.ID)
	if err != nil {
		http.Error(w, "groupedEvents:"+err.Error(), http.StatusInternalServerError)
		return
	}

	// queries := yikes.New(db.Conn)
	// tasks, err := queries.FindTasksByUserID(ctx, user.ID)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	err = pageTemplate.Execute(w, struct {
		GroupedEvents google.TimeGrouping
		//		Tasks         []yikes.Task
	}{
		GroupedEvents: *groupedEvents,
		//Tasks:         tasks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
