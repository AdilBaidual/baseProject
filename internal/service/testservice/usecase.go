package testservice

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// TestRepository defines the interface for test-related operations
type TestRepository interface {
	Pong(ctx context.Context) (string, error)
}

// Service handles business logic for test operations
type Service struct {
	logger *zap.Logger
	repo   TestRepository
}

// NewService creates a new test service instance
func NewService(logger *zap.Logger, repo TestRepository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

// Pong handles the pong business logic
func (s *Service) Pong(ctx context.Context) (string, error) {
	s.logger.Info("Processing pong request")

	result, err := s.repo.Pong(ctx)
	if err != nil {
		s.logger.Error("Failed to execute pong operation", zap.Error(err))
		return "", fmt.Errorf("pong operation failed: %w", err)
	}

	s.logger.Debug("Pong operation completed successfully", zap.String("result", result))
	return result, nil
}
