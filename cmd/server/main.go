package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/a2ikm/keflavik/model"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
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

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/keflavik?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect postgres: %v", err)
	}

	queries := model.New(db)

	mux := http.NewServeMux()
	mux.Handle("/authenticate", &authenticateHandler{queries})
	mux.Handle("/create_user", &createUserHandler{queries})
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
	const uniquenessViolation = pq.ErrorCode("23505")
	if pgerr, ok := err.(*pq.Error); ok {
		if pgerr.Code == uniquenessViolation {
			return true
		}
	}
	return false
}
