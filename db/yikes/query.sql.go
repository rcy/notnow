// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: query.sql

package yikes

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createSession = `-- name: CreateSession :one
insert into sessions(user_id) values($1) returning id
`

func (q *Queries) CreateSession(ctx context.Context, userID pgtype.UUID) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createSession, userID)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const createToken = `-- name: CreateToken :one
insert into tokens(token, user_id) values($1, $2) returning id
`

type CreateTokenParams struct {
	Token  []byte
	UserID pgtype.UUID
}

func (q *Queries) CreateToken(ctx context.Context, arg CreateTokenParams) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, createToken, arg.Token, arg.UserID)
	var id pgtype.UUID
	err := row.Scan(&id)
	return id, err
}

const createUser = `-- name: CreateUser :one
insert into users(email) values($1) returning id, created_at, email
`

func (q *Queries) CreateUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, createUser, email)
	var i User
	err := row.Scan(&i.ID, &i.CreatedAt, &i.Email)
	return i, err
}

const findTokenByUserID = `-- name: FindTokenByUserID :one
select tokens.id, tokens.created_at, tokens.token, tokens.user_id from tokens join users on tokens.user_id = users.id where users.id = $1 limit 1
`

func (q *Queries) FindTokenByUserID(ctx context.Context, id pgtype.UUID) (Token, error) {
	row := q.db.QueryRow(ctx, findTokenByUserID, id)
	var i Token
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.Token,
		&i.UserID,
	)
	return i, err
}

const findUserByEmail = `-- name: FindUserByEmail :one
select id, created_at, email from users where email = $1
`

func (q *Queries) FindUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, findUserByEmail, email)
	var i User
	err := row.Scan(&i.ID, &i.CreatedAt, &i.Email)
	return i, err
}

const findUserBySessionID = `-- name: FindUserBySessionID :one
select users.id, users.created_at, users.email from sessions join users on sessions.user_id = users.id where sessions.id = $1
`

func (q *Queries) FindUserBySessionID(ctx context.Context, id pgtype.UUID) (User, error) {
	row := q.db.QueryRow(ctx, findUserBySessionID, id)
	var i User
	err := row.Scan(&i.ID, &i.CreatedAt, &i.Email)
	return i, err
}
