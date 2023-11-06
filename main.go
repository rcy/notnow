package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"yikes/config"
	"yikes/db/yikes"
	ymw "yikes/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
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

	r.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "conn", conn)))
			})
		})

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
				TimeMax(time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)).
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

	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.GoogleOAuth2.AuthCodeURL("", oauth2.AccessTypeOffline, oauth2.ApprovalForce), http.StatusFound)
	})

	r.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := yikes.New(conn)

		code := r.URL.Query().Get("code")
		token, err := config.GoogleOAuth2.Exchange(ctx, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := config.GoogleOAuth2.Client(ctx, token)

		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo := struct {
			Email string `json:"email"`
		}{}
		json.Unmarshal(bytes, &userInfo)

		user, err := queries.FindUserByEmail(ctx, userInfo.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				user, err = queries.CreateUser(ctx, userInfo.Email)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if user.Email == "" {
			panic("EMPTY EMAIL")
		}

		bytes, err = json.Marshal(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = queries.CreateToken(ctx, yikes.CreateTokenParams{Token: bytes, UserID: user.ID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		sessionID, err := queries.CreateSession(ctx, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		value, err := sessionID.Value()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "yikes.session",
			Value:   value.(string),
			Expires: time.Now().Add(30 * 24 * time.Hour),
			Path:    "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
