package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/a2ikm/keflavik/model"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
)

type createUserRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type createUserResponse struct {
	Ok   bool `json:"ok"`
	Data struct {
		Name string `json:"name"`
	} `json:"data"`
}

type createUserHandler struct {
	queries *model.Queries
}

func (h *createUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "bad_request", "Failed to parse request: %v", err)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErrorResponse(w, "internal_server_error", "Failed to generate hashed password: %v", err)
		return
	}

	params := model.CreateUserParams{
		Name:         req.Name,
		PasswordHash: string(hashed),
	}
	if err := h.queries.CreateUser(r.Context(), params); err != nil {
		if isUniquenessViolation(err) {
			writeErrorResponse(w, "bad_request", "name is already taken")
			return
		} else {
			writeErrorResponse(w, "internal_server_error", "Failed to store user information: %v", err)
			return
		}
	}

	res := createUserResponse{
		Ok: true,
		Data: struct {
			Name string `json:"name"`
		}{
			Name: req.Name,
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

type authenticateRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type authenticateResponse struct {
	Ok   bool `json:"ok"`
	Data struct {
		Name        string `json:"name"`
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type authenticateHandler struct {
	queries *model.Queries
}

func (h *authenticateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var req authenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "bad_request", "Failed to parse request: %v", err)
		return
	}

	user, err := h.queries.GetUserByName(r.Context(), req.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			writeErrorResponse(w, "unauthorized", "name or password is incorrect")
			return
		} else {
			writeErrorResponse(w, "internal_server_error", "Failed to fetch user: %v", err)
			return
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeErrorResponse(w, "unauthorized", "name or password is incorrect")
		return
	}

	var session model.Session
	for {
		token, err := generateRandomString(64)
		if err != nil {
			continue
		}

		params := model.CreateSessionParams{
			UserID:      user.ID,
			AccessToken: token,
		}
		session, err = h.queries.CreateSession(r.Context(), params)
		if err != nil {
			if isUniquenessViolation(err) {
				continue
			}
			writeErrorResponse(w, "internal_server_error", "Failed to fetch user: %v", err)
			return
		}

		break
	}

	res := authenticateResponse{
		Ok: true,
		Data: struct {
			Name        string `json:"name"`
			AccessToken string `json:"access_token"`
		}{
			Name:        user.Name,
			AccessToken: session.AccessToken,
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

type createPostRequest struct {
	Body string `json:"body"`
}

type createPostResponse struct {
	Ok   bool `json:"ok"`
	Data struct {
	} `json:"data"`
}

type createPostHandler struct {
	queries *model.Queries
}

func (h *createPostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	user, err := authenticateWithAccessToken(h.queries, r)
	if err != nil {
		writeErrorResponse(w, "unauthorized", "Failed to authenticate: %v", err)
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "bad_request", "Failed to parse request: %v", err)
		return
	}

	params := model.CreatePostParams{
		UserID:    user.ID,
		Body:      req.Body,
		CreatedAt: time.Now().UTC(),
	}
	_, err = h.queries.CreatePost(r.Context(), params)
	if err != nil {
		writeErrorResponse(w, "bad_request", "Failed to store post: %v", err)
		return
	}

	res := createPostResponse{
		Ok:   true,
		Data: struct{}{},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

type PostInResponse struct {
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type getPostsResponse struct {
	Ok   bool `json:"ok"`
	Data struct {
		Posts []PostInResponse `json:"posts"`
	} `json:"data"`
}

type getPostsHandler struct {
	queries *model.Queries
}

func (h *getPostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	user, err := authenticateWithAccessToken(h.queries, r)
	if err != nil {
		writeErrorResponse(w, "unauthorized", "Failed to authenticate: %v", err)
		return
	}

	posts, err := h.queries.GetPostsByUserId(r.Context(), user.ID)
	if err != nil {
		writeErrorResponse(w, "bad_request", "Failed to store post: %v", err)
		return
	}

	postsInResponse := make([]PostInResponse, len(posts))
	for _, post := range posts {
		postsInResponse = append(postsInResponse, PostInResponse{
			Body:      post.Body,
			CreatedAt: post.CreatedAt,
		})
	}

	res := getPostsResponse{
		Ok: true,
		Data: struct {
			Posts []PostInResponse `json:"posts"`
		}{
			Posts: postsInResponse,
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5432/keflavik?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect postgres: %v", err)
	}
	defer conn.Close(context.Background())

	queries := model.New(conn)

	mux := http.NewServeMux()
	mux.Handle("/authenticate", &authenticateHandler{queries})
	mux.Handle("/create_user", &createUserHandler{queries})
	mux.Handle("/create_post", &createPostHandler{queries})
	mux.Handle("/get_posts", &getPostsHandler{queries})
	mux.Handle("/", http.NotFoundHandler())

	log.Printf("Start listening on :8080")
	if err := http.ListenAndServe(":8000", mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %v", err)
	}
}

func generateRandomString(digit uint32) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, digit)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	var result string
	for _, v := range b {
		result += string(letters[int(v)%len(letters)])
	}
	return result, nil
}

func isUniquenessViolation(err error) bool {
	pgerr, ok := err.(*pgconn.PgError)
	if !ok {
		return false
	}

	return pgerr.Code == "23505"
}

func authenticateWithAccessToken(queries *model.Queries, r *http.Request) (model.User, error) {
	authorization := r.Header.Get("Authorization")
	if len(authorization) == 0 {
		return model.User{}, fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(authorization, " ")
	if len(parts) != 2 {
		return model.User{}, fmt.Errorf("malformed Authorization header")
	}
	if strings.ToLower(parts[0]) != "bearer" {
		return model.User{}, fmt.Errorf("malformed Authorization header")
	}

	session, err := queries.GetSessionByAccessToken(r.Context(), parts[1])
	if err != nil {
		if err == sql.ErrNoRows {
			return model.User{}, fmt.Errorf("wrong access token")
		} else {
			return model.User{}, fmt.Errorf("something wrong")
		}
	}

	user, err := queries.GetUserById(r.Context(), session.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.User{}, fmt.Errorf("user associated with access token not found")
		} else {
			return model.User{}, fmt.Errorf("something wrong")
		}
	}

	return user, nil
}

type errorResponse struct {
	Ok           bool   `json:"ok"`
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

func writeErrorResponse(w http.ResponseWriter, errorCode string, format string, a ...any) {
	res := errorResponse{
		Ok:           false,
		ErrorCode:    errorCode,
		ErrorMessage: fmt.Sprintf(format, a...),
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
