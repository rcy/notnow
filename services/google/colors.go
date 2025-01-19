package google

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/api/calendar/v3"
)

func Colors(ctx context.Context, userID pgtype.UUID) (*calendar.Colors, error) {
	srv, err := ServiceForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	c, err := srv.Colors.Get().Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return c, nil
}
