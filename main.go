package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type usersHandler struct{}

func (h *usersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "bad_request", "Failed to parse request: %v", err)
		return
	}

	res := createUserResponse{
		Ok: true,
		Data: struct {
			Name string "json:\"name\""
		}{
			Name: req.Name,
		},
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/users", &usersHandler{})
	mux.Handle("/", http.NotFoundHandler())

	log.Printf("Start listening on :8080")
	if err := http.ListenAndServe(":8000", mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %v", err)
	}
}
