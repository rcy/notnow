package middleware

import (
	"context"
	"errors"
	"net/http"
	"yikes/db"
	"yikes/db/yikes"
	"yikes/routes/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func User(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		queries := yikes.New(db.Conn)

		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
			return
		}
		uuid := pgtype.UUID{}
		err = uuid.Scan(cookie.Value)
		if err != nil {
			http.Error(w, "uuid scan:"+err.Error(), http.StatusInternalServerError)
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
