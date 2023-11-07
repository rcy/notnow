package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"yikes/config"
	"yikes/db"
	"yikes/db/yikes"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
)

func User(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		queries := yikes.New(db.Conn)

		cookie, err := r.Cookie("yikes.session")
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		uuid := pgtype.UUID{}
		err = uuid.Scan(cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := queries.FindUserBySessionID(ctx, uuid)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Redirect(w, r, "/auth", http.StatusSeeOther)
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "user", user)))
	})
}

func UserFromContext(ctx context.Context) (yikes.User, bool) {
	value, ok := ctx.Value("user").(yikes.User)
	return value, ok
}

func Client(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, _ := UserFromContext(ctx)

		queries := yikes.New(db.Conn)

		token, err := queries.FindTokenByUserID(ctx, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		oauthToken := oauth2.Token{}
		err = json.Unmarshal(token.Token, &oauthToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := config.GoogleOAuth2.Client(ctx, &oauthToken)

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, "client", client)))
	})
}

func ClientFromContext(ctx context.Context) (*http.Client, bool) {
	value, ok := ctx.Value("client").(*http.Client)
	return value, ok
}
