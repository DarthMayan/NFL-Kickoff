package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
	Active    bool      `json:"active"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullName,omitempty"`
	Active   *bool  `json:"active,omitempty"`
}
