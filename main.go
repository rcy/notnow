package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"yikes/db/yikes"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

var (
	ClientID     = os.Getenv("GOOGLE_CLIENT_ID")
	ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
)

var oauthConfig = oauth2.Config{
	ClientID:     ClientID,
	ClientSecret: ClientSecret,
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://accounts.google.com/o/oauth2/token",
	},
	RedirectURL: "http://localhost:8080/auth/callback",
	Scopes: []string{
		"openid",
		"profile",
		"email",
		"https://www.googleapis.com/auth/calendar.events",
		"https://www.googleapis.com/auth/calendar.readonly",
	},
}

func main() {
	ctx := context.TODO()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	//r.Route("/", app.Router)

	r.Get("/auth", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, oauthConfig.AuthCodeURL("", oauth2.AccessTypeOffline, oauth2.ApprovalForce), http.StatusFound)
	})

	r.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := oauthConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		queries := yikes.New(conn)

		key, err := queries.CreateToken(context.TODO(), bytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// key, err := tokens.Store(*token)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		http.SetCookie(w, &http.Cookie{
			Name:    "ike-session",
			Value:   key,
			Expires: time.Now().Add(30 * 24 * time.Hour),
			Path:    "/",
		})

		// client := oauthConfig.Client(context.Background(), token)
		// srv, err := calendar.New(client)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// Now, you can use 'srv' to make API calls to Google Calendar.
		// For example, list the user's calendars:
		// calendars, err := srv.CalendarList.List().Do()
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// events, err := srv.Events.
		// 	List("primary").
		// 	TimeMin(time.Now().Format(time.RFC3339)).
		// 	TimeMax(time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)).
		// 	Do()

		// for _, e := range events.Items {
		// 	log.Printf("* %v\n", e)
		// }

		_, err = w.Write([]byte("Authenticated successfully!"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
