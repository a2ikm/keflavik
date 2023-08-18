-- name: CreateUser :exec
INSERT INTO users (name, password_hash) VALUES ($1, $2);

-- name: GetUserByName :one
SELECT name, password_hash FROM users WHERE name = $1 LIMIT 1;
