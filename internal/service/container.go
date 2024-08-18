package service

import (
	"github.com/AdilBaidual/baseProject/internal/service/test_service"
	"github.com/AdilBaidual/baseProject/internal/store"
	"go.uber.org/zap"
)

type ServiceContainer struct {
	testService *test_service.Service
}

func NewServiceContainer(logger *zap.Logger, testStore *store.Store) *ServiceContainer {
	return &ServiceContainer{
		testService: test_service.NewService(logger, testStore),
	}
}

func (s *ServiceContainer) GetTestService() *test_service.Service {
	return s.testService
}
