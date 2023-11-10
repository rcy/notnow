-- name: FindTokenByUserID :one
select tokens.* from tokens join users on tokens.user_id = users.id where users.id = $1 limit 1;

-- name: CreateToken :one
insert into tokens(token, user_id) values($1, $2) returning id;

-- name: CreateUser :one
insert into users(email) values($1) returning *;

-- name: FindUserByEmail :one
select * from users where email = $1;

-- name: CreateSession :one
insert into sessions(user_id) values($1) returning id;

-- name: FindUserBySessionID :one
select users.* from sessions join users on sessions.user_id = users.id where sessions.id = $1;

-- name: FindTasksByUserID :many
select * from tasks where user_id = $1 order by created_at desc limit 1000;

-- name: FindTaskByEventID :one
select tasks.* from tasks
join task_events on tasks.id = task_events.task_id
where task_events.event_id = $1
limit 1;

-- name: UserTaskByID :one
select * from tasks where user_id = $1 and id = $2;

-- name: CreateTask :one
insert into tasks(summary, user_id) values($1, $2) returning *;

-- name: CreateUserTaskEvent :one
insert into task_events(user_id, task_id, event_id) values($1, $2, $3) returning *;
