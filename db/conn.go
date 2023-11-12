package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Conn *pgxpool.Pool

func MustConnect(ctx context.Context, connString string) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		panic(err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		panic(err)
	}
	Conn = pool
}
