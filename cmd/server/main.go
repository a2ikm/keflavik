package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
	db *sql.DB
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

	if _, err := h.db.Exec("INSERT INTO users (name, password_hash) VALUES ($1, $2);", req.Name, hashed); err != nil {
		writeErrorResponse(w, "internal_server_error", "Failed to store user information: %v", err)
		return
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
		Name string `json:"name"`
	} `json:"data"`
}

type authenticateHandler struct {
	db *sql.DB
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

	rows, err := h.db.Query("SELECT name, password_hash FROM users WHERE name = $1 LIMIT 1", req.Name)
	if err != nil {
		writeErrorResponse(w, "internal_server_error", "Failed to fetch user: %v", err)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeErrorResponse(w, "unauthorized", "name or password is incorrect")
		return
	}

	var name string
	var passwordHash string
	if err = rows.Scan(&name, &passwordHash); err != nil {
		writeErrorResponse(w, "internal_server_error", "Failed to scan row: %v", err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		writeErrorResponse(w, "unauthorized", "name or password is incorrect")
		return
	}

	res := authenticateResponse{
		Ok: true,
		Data: struct {
			Name string `json:"name"`
		}{
			Name: name,
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

	mux := http.NewServeMux()
	mux.Handle("/authenticate", &authenticateHandler{db})
	mux.Handle("/create_user", &createUserHandler{db})
	mux.Handle("/", http.NotFoundHandler())

	log.Printf("Start listening on :8080")
	if err := http.ListenAndServe(":8000", mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %v", err)
	}
}
