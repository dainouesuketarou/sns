-- name: CreateUser :one
insert into users (id, username, description, avatar_url, created_at)
values ($1, $2, $3, $4, $5)
returning *;

-- name: GetUserByID :one
select * from users
where id = $1;
