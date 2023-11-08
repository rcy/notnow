package google

import (
	"context"
	"encoding/json"
	"net/http"
	"yikes/db"
	"yikes/db/yikes"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
)

func clientForUser(ctx context.Context, userID pgtype.UUID) (*http.Client, error) {
	queries := yikes.New(db.Conn)

	token, err := queries.FindTokenByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	oauthToken := oauth2.Token{}
	err = json.Unmarshal(token.Token, &oauthToken)
	if err != nil {
		return nil, err
	}

	return Config.Client(ctx, &oauthToken), nil
}

func ServiceForUser(ctx context.Context, userID pgtype.UUID) (*calendar.Service, error) {
	client, err := clientForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return calendar.New(client)
}
