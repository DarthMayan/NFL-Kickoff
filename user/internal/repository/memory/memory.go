package memory

import (
	"context"
	"os/user"
	"sync"
	"time"

	"kickoff.com/user/internal/repository"
)

type Repository struct {
	sync.RWMutex
	users map[string]*user.User
}

func New() *Repository {
	return &Repository{
		users: make(map[string]*user.User),
	}
}

func (r *Repository) Get(ctx context.Context, id string) (*user.User, error) {
	r.RLock()
	defer r.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	r.RLock()
	defer r.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (r *Repository) Put(ctx context.Context, id string, u *user.User) error {
	r.Lock()
	defer r.Unlock()

	u.CreatedAt = time.Now()
	u.Active = true
	r.users[id] = u
	return nil
}

func (r *Repository) GetAll(ctx context.Context) ([]*user.User, error) {
	r.RLock()
	defer r.RUnlock()

	var users []*user.User
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}
