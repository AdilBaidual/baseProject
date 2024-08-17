package usecase

import (
	test "github.com/AdilBaidual/baseProject/internal/TestEntity"
	"go.uber.org/zap"
)

type TestUseCase struct {
	logger *zap.Logger
}

func NewTestUseCase(logger *zap.Logger) test.UseCase {
	return &TestUseCase{
		logger: logger,
	}
}
