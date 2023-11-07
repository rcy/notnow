package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var Conn *pgx.Conn

func MustConnect(ctx context.Context, connString string) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		panic(err)
	}
	Conn = conn
}
