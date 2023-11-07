package tasks

import (
	_ "embed"
	"html/template"
	"net/http"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/middleware"

	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router) {
	r.Post("/", post)
}

var (
	//go:embed partials.gohtml
	Content  string
	partials = template.Must(template.New("tasks").Parse(Content))
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
	_, err := queries.CreateTask(ctx, yikes.CreateTaskParams{
		Summary: summary,
		UserID:  user.ID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tasks, err := queries.FindTasksByUserID(ctx, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	partials.ExecuteTemplate(w, "tasks/section", tasks)
}
