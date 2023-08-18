-- name: CreateUser :exec
INSERT INTO users (name, password_hash) VALUES ($1, $2);

-- name: GetUserById :one
SELECT id, name, password_hash FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByName :one
SELECT id, name, password_hash FROM users WHERE name = $1 LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, access_token) VALUES ($1, $2) RETURNING *;

-- name: GetSessionByAccessToken :one
SELECT id, user_id, access_token FROM sessions WHERE access_token = $1 LIMIT 1;

-- name: CreatePost :one
INSERT INTO posts (user_id, body, created_at) VALUES ($1, $2, $3) RETURNING *;

-- name: GetPostsByUserId :many
SELECT id, user_id, body, created_at FROM posts WHERE user_id = $1;
