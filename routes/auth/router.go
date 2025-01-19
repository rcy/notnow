package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/internal/html"
	"yikes/services/google"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const CookieName = "notnow.session"

func Router(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		err := html.Layout("auth",
			Div(Text("not now")),
			A(Href("/auth/google"), Text("login"), Class(html.ButtonClass)),
		).Render(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	r.Get("/google", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, google.Config.AuthCodeURL("", oauth2.AccessTypeOffline, oauth2.ApprovalForce), http.StatusFound)
	})

	r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		queries := yikes.New(db.Conn)

		code := r.URL.Query().Get("code")
		token, err := google.Config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := google.Config.Client(ctx, token)

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
			Name:    CookieName,
			Value:   value.(string),
			Expires: time.Now().Add(30 * 24 * time.Hour),
			Path:    "/",
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
