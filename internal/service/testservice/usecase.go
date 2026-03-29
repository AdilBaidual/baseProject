package testservice

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type TestRepository interface {
	Pong(ctx context.Context) (string, error)
}

type Service struct {
	logger *zap.Logger
	repo   TestRepository
}

func NewService(logger *zap.Logger, repo TestRepository) *Service {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

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
