-- name: GetTokenBySessionKey :one
select * from tokens where session_key = $1 limit 1;

-- name: CreateToken :one
insert into tokens(token) values($1) returning uuid::text;
