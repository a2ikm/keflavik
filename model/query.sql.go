// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: query.sql

package model

import (
	"context"
	"time"
)

const createPost = `-- name: CreatePost :one
INSERT INTO posts (user_id, body, created_at) VALUES ($1, $2, $3) RETURNING id, user_id, body, created_at
`

type CreatePostParams struct {
	UserID    int32
	Body      string
	CreatedAt time.Time
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, createPost, arg.UserID, arg.Body, arg.CreatedAt)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Body,
		&i.CreatedAt,
	)
	return i, err
}

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (user_id, access_token) VALUES ($1, $2) RETURNING id, user_id, access_token
`

type CreateSessionParams struct {
	UserID      int32
	AccessToken string
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRowContext(ctx, createSession, arg.UserID, arg.AccessToken)
	var i Session
	err := row.Scan(&i.ID, &i.UserID, &i.AccessToken)
	return i, err
}

const createUser = `-- name: CreateUser :exec
INSERT INTO users (name, password_hash) VALUES ($1, $2)
`

type CreateUserParams struct {
	Name         string
	PasswordHash string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.ExecContext(ctx, createUser, arg.Name, arg.PasswordHash)
	return err
}

const getPostsByUserId = `-- name: GetPostsByUserId :many
SELECT id, user_id, body, created_at FROM posts WHERE user_id = $1
`

func (q *Queries) GetPostsByUserId(ctx context.Context, userID int32) ([]Post, error) {
	rows, err := q.db.QueryContext(ctx, getPostsByUserId, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Post
	for rows.Next() {
		var i Post
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Body,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSessionByAccessToken = `-- name: GetSessionByAccessToken :one
SELECT id, user_id, access_token FROM sessions WHERE access_token = $1 LIMIT 1
`

func (q *Queries) GetSessionByAccessToken(ctx context.Context, accessToken string) (Session, error) {
	row := q.db.QueryRowContext(ctx, getSessionByAccessToken, accessToken)
	var i Session
	err := row.Scan(&i.ID, &i.UserID, &i.AccessToken)
	return i, err
}

const getUserById = `-- name: GetUserById :one
SELECT id, name, password_hash FROM users WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserById(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserById, id)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.PasswordHash)
	return i, err
}

const getUserByName = `-- name: GetUserByName :one
SELECT id, name, password_hash FROM users WHERE name = $1 LIMIT 1
`

func (q *Queries) GetUserByName(ctx context.Context, name string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByName, name)
	var i User
	err := row.Scan(&i.ID, &i.Name, &i.PasswordHash)
	return i, err
}
