package controller

import (
	"context"
	"errors"
	"fmt"
	"os/user"

	"kickoff.com/user/internal/repository"
)

type userRepository interface {
	Get(ctx context.Context, id string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	Put(ctx context.Context, id string, user *user.User) error
	GetAll(ctx context.Context) ([]*user.User, error)
}

type Controller struct {
	repo userRepository
}

func New(repo userRepository) Controller {
	return Controller{repo: repo}
}

func (c Controller) Get(ctx context.Context, id string) (*user.User, error) {
	res, err := c.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c Controller) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	res, err := c.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c Controller) CreateUser(ctx context.Context, username, email, fullName string) (*user.User, error) {
	// Validar que el username no exista ya
	_, err := c.repo.GetByUsername(ctx, username)
	if err == nil {
		return nil, errors.New("username already exists")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	// Validaciones básicas
	if username == "" || email == "" || fullName == "" {
		return nil, errors.New("username, email and fullName are required")
	}

	// Crear el usuario
	userID := fmt.Sprintf("user_%s", username)
	newUser := &user.User{
		ID:       userID,
		Username: username,
		Email:    email,
		FullName: fullName,
	}

	if err := c.repo.Put(ctx, userID, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

func (c Controller) GetAllUsers(ctx context.Context) ([]*user.User, error) {
	return c.repo.GetAll(ctx)
}
