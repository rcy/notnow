package main

import (
	"context"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"yikes/routes/auth"

	ymw "yikes/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"google.golang.org/api/calendar/v3"
)

var (
	//go:embed page.gohtml
	pageContent string

	pageTemplate = template.Must(template.New("page").Parse(pageContent))
)

func main() {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "conn", conn)))
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(ymw.User)
		r.Use(ymw.Client)

		r.Get("/user", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, ok := ymw.UserFromContext(ctx)
			if !ok {
				http.Error(w, "couldn't get user", http.StatusInternalServerError)
				return
			}

			s, err := user.ID.Value()
			log.Print(s, err)
		})

		r.Get("/cal", func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			client, ok := ymw.ClientFromContext(ctx)
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
		})
	})

	r.Route("/auth", auth.Router)

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
