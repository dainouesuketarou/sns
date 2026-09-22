-- name: CreatePost :one
insert into posts (id, user_id, body, image_url, is_public, created_at)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: GetPostByID :one
select * from posts
where id = $1;
