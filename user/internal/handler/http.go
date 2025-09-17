package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os/user"
	"strings"

	"kickoff.com/user/internal/repository"
)

type userController interface {
	Get(ctx context.Context, id string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	CreateUser(ctx context.Context, username, email, fullName string) (*user.User, error)
	GetAllUsers(ctx context.Context) ([]*user.User, error)
}

type Handler struct {
	ctrl userController
}

func New(ctrl userController) Handler {
	return Handler{ctrl: ctrl}
}

func (h Handler) GetUser(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	username := req.URL.Query().Get("username")

	if id == "" && username == "" {
		http.Error(w, "id or username parameter is required", http.StatusBadRequest)
		return
	}

	var u *user.User
	var err error

	if id != "" {
		u, err = h.ctrl.Get(req.Context(), id)
	} else {
		u, err = h.ctrl.GetByUsername(req.Context(), username)
	}

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("Repository error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Printf("Response encode error: %v", err)
	}
}

func (h Handler) CreateUser(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type createUserRequest struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"fullName"`
	}

	var createReq createUserRequest
	if err := json.NewDecoder(req.Body).Decode(&createReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	u, err := h.ctrl.CreateUser(req.Context(), createReq.Username, createReq.Email, createReq.FullName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "required") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Controller error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(u); err != nil {
		log.Printf("Response encode error: %v", err)
	}
}

func (h Handler) GetAllUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.ctrl.GetAllUsers(req.Context())
	if err != nil {
		log.Printf("Controller error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users": users,
		"total": len(users),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Response encode error: %v", err)
	}
}
