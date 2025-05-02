package fakes

import (
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/users"
	"context"
	"errors"
	"sync"
	"time"
)

type FakeUserRepository struct {
	mu           sync.Mutex
	usersByEmail map[string]*users.User
	usersByID    map[int64]*users.User
	nextID       int64
}

// NewFakeUserRepository creates a new fake user repository.
func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{
		usersByEmail: make(map[string]*users.User),
		usersByID:    make(map[int64]*users.User),
		nextID:       1,
	}
}

// Create inserts a new user into the fake repository.
func (r *FakeUserRepository) Create(ctx context.Context, dto *user_dto.CreateUserDTO) (*users.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.usersByEmail[dto.Email]; exists {
		return nil, errors.New("user already exists")
	}

	newUser := &users.User{
		ID:        users.UserId(r.nextID),
		Username:  dto.Name,
		Email:     dto.Email,
		Password:  dto.Password, // In a real flow, the password is already hashed by the auth service.
		CreatedAt: time.Now(),
	}
	r.usersByEmail[dto.Email] = newUser
	r.usersByID[r.nextID] = newUser
	r.nextID++
	return newUser, nil
}

// GetUserByEmail retrieves a user by email.
func (r *FakeUserRepository) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, ok := r.usersByEmail[email]; ok {
		return user, nil
	}
	return nil, nil
}

// (Optional) GetUserById retrieves a user by ID.
func (r *FakeUserRepository) GetUserById(ctx context.Context, id users.UserId) (*users.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user, ok := r.usersByID[int64(id)]; ok {
		return user, nil
	}
	return nil, nil
}
