package test_service

import (
	"go.uber.org/zap"
)

type testStore interface {
	Pong() string
}

type Service struct {
	logger *zap.Logger

	testStore testStore
}

func NewService(logger *zap.Logger, testStore testStore) *Service {
	return &Service{
		logger:    logger,
		testStore: testStore,
	}
}

func (s *Service) Pong() string {
	return s.testStore.Pong()
}
