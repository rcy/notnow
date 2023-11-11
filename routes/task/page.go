package task

import (
	_ "embed"
	"html/template"
	"net/http"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/layout"
	"yikes/middleware"
	"yikes/routes/events"
	"yikes/services/google"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

var (
	//go:embed page.gohtml
	pageContent  string
	pageTemplate = template.Must(template.New("task").Funcs(events.Funcs).Parse(layout.Content + pageContent))
)

func Router(r chi.Router) {
	r.Get("/{eventID}/{state}", page)
	r.Post("/{eventID}/done", postDone)
	r.Post("/{eventID}/excuse", postExcuse)
}

func page(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := yikes.New(db.Conn)

	task, err := queries.FindTaskByEventID(ctx, chi.URLParam(r, "eventID"))
	if err != nil {
		if err == pgx.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageTemplate.Execute(w, struct {
		State string
		Task  yikes.Task
	}{
		State: chi.URLParam(r, "state"),
		Task:  task,
	})
}

func postDone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, _ := middleware.UserFromContext(ctx)
	eventID := chi.URLParam(r, "eventID")

	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(ctx)

	queries := yikes.New(tx)
	task, err := queries.FindTaskByEventID(ctx, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = queries.SetTaskStatus(ctx, yikes.SetTaskStatusParams{ID: task.ID, Status: "done"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = google.DeleteEvent(ctx, user.ID, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func postExcuse(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
