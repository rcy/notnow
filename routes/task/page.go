package task

import (
	_ "embed"
	"html/template"
	"net/http"
	"strings"
	"yikes/layout"
	"yikes/middleware"
	"yikes/routes/events"
	"yikes/services/google"

	"github.com/go-chi/chi/v5"
)

var (
	//go:embed page.gohtml
	pageContent  string
	pageTemplate = template.Must(template.New("task").Funcs(events.Funcs).Parse(layout.Content + pageContent))
)

func Router(r chi.Router) {
	r.Get("/{eventID}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.URL.String()+"/show", http.StatusSeeOther)
	})
	r.Get("/{eventID}/{state}", page)
	r.Post("/{eventID}/done", postDone)
	r.Post("/{eventID}/excuse", postExcuse)
	r.Post("/{eventID}/summary", postSummary)
}

func page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := middleware.UserFromContext(ctx)
	eventID := chi.URLParam(r, "eventID")
	state := chi.URLParam(r, "state")
	if state == "" {
		state = "show"
	}
	//queries := yikes.New(db.Conn)

	// task, err := queries.FindTaskByEventID(ctx, chi.URLParam(r, "eventID"))
	// if err != nil {
	// 	if err == pgx.ErrNoRows {
	// 		http.NotFound(w, r)
	// 		return
	// 	}
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	event, err := google.UserEvent(ctx, user.ID, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageTemplate.Execute(w, struct {
		State string
		Event *google.EventModel
	}{
		State: state,
		//Task:  task,
		Event: event,
	})
}

func postDone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := middleware.UserFromContext(ctx)
	eventID := chi.URLParam(r, "eventID")

	// tx, err := db.Conn.Begin(ctx)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// defer tx.Rollback(ctx)

	// queries := yikes.New(tx)
	// task, err := queries.FindTaskByEventID(ctx, eventID)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// err = queries.SetTaskStatus(ctx, yikes.SetTaskStatusParams{ID: task.ID, Status: "done"})
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	err := google.DeleteEvent(ctx, user.ID, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// err = tx.Commit(ctx)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func postSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := middleware.UserFromContext(ctx)
	eventID := chi.URLParam(r, "eventID")
	summary := strings.TrimSpace(r.FormValue("summary"))
	if summary == "" {
		http.Error(w, "empty summary", http.StatusBadRequest)
		return
	}

	//queries := yikes.New(db.Conn)

	// task, err := queries.FindTaskByEventID(ctx, eventID)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// task, err = queries.SetTaskSummary(ctx, yikes.SetTaskSummaryParams{ID: task.ID, Summary: summary})
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	event, err := google.UpdateEventSummary(ctx, user.ID, eventID, summary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = pageTemplate.ExecuteTemplate(w, "title", event)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postExcuse(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
