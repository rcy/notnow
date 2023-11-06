-- name: GetTokenBySessionKey :one
select * from tokens where session_key = $1 limit 1;

-- name: CreateToken :exec
insert into tokens(token, session_key) values($1,$2);
