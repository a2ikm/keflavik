-- name: CreateUser :exec
INSERT INTO users (name, password_hash) VALUES ($1, $2);

-- name: GetUserByName :one
SELECT id, name, password_hash FROM users WHERE name = $1 LIMIT 1;

-- name: CreateSession :one
INSERT INTO sessions (user_id, access_token) VALUES ($1, $2) RETURNING *;
