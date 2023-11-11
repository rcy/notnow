package tasks

import (
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/middleware"
	"yikes/routes/events"
	"yikes/services/google"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func Router(r chi.Router) {
	r.Post("/", post)
	r.Post("/{id}/schedule", postSchedule)
}

var (
	//go:embed partials.gohtml
	Content  string
	partials = template.Must(template.New("tasks").Funcs(events.Funcs).Parse(Content))
)

func post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary := r.FormValue("summary")
	if summary == "" {
		http.Error(w, "empty summary", http.StatusBadRequest)
		return
	}

	user, _ := middleware.UserFromContext(ctx)

	queries := yikes.New(db.Conn)
	task, err := queries.CreateTask(ctx, yikes.CreateTaskParams{
		Summary: summary,
		UserID:  user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = scheduleTask(ctx, user.ID, task.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", "calendarUpdated")

	partials.ExecuteTemplate(w, "tasks/section", nil)
}

func postSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	taskID := pgtype.UUID{}
	err := taskID.Scan(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, _ := middleware.UserFromContext(ctx)

	err = scheduleTask(ctx, user.ID, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", "calendarUpdated")

	w.WriteHeader(http.StatusCreated)
}

func scheduleTask(ctx context.Context, userID pgtype.UUID, taskID pgtype.UUID) error {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := yikes.New(tx)

	task, err := queries.UserTaskByID(ctx, yikes.UserTaskByIDParams{UserID: userID, ID: taskID})
	if err != nil {
		return err
	}

	event, err := google.CreateTaskEvent(ctx, userID, task)

	_, err = queries.CreateUserTaskEvent(ctx, yikes.CreateUserTaskEventParams{UserID: userID, TaskID: taskID, EventID: event.Id})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
