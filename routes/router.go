package routes

import (
	_ "embed"
	"yikes/middleware"
	"yikes/routes/auth"
	"yikes/routes/tasks"

	"github.com/go-chi/chi/v5"
)

func Router(r chi.Router) {
	r.Route("/auth", auth.Router)

	r.Group(func(r chi.Router) {
		r.Use(middleware.User)
		r.Get("/", Page)
		r.Route("/tasks", tasks.Router)
	})
}
