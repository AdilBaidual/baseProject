package store

import (
	"context"
)

type TestRepository interface {
	Pong(ctx context.Context) (string, error)
}

type Repository struct {
	Test TestRepository
}

func NewRepository(test TestRepository) *Repository {
	return &Repository{
		Test: test,
	}
}
