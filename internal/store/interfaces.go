package store

import (
	"context"
)

// TestRepository defines the interface for test-related database operations
type TestRepository interface {
	Pong(ctx context.Context) (string, error)
	// Add more methods as needed
	// GetUserByID(ctx context.Context, id string) (*User, error)
	// CreateUser(ctx context.Context, user *User) error
}

// Repository aggregates all repository interfaces
type Repository struct {
	Test TestRepository
}

// NewRepository creates a new Repository instance
func NewRepository(test TestRepository) *Repository {
	return &Repository{
		Test: test,
	}
}
