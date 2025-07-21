package test

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/AdilBaidual/baseProject/internal/pb/baseProject/test"
)

// Service defines the interface for test service operations
type Service interface {
	Pong(ctx context.Context) (string, error)
}

// Handler handles gRPC requests for test service
type Handler struct {
	test.TestServiceServer

	logger  *zap.Logger
	service Service
}

// NewHandler creates a new test handler instance
func NewHandler(logger *zap.Logger, service Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// Register registers the handler with gRPC server and HTTP gateway
func Register(mux *runtime.ServeMux, conn *grpc.ClientConn, gRPCServer *grpc.Server, handler *Handler) error {
	// Register with gRPC server
	test.RegisterTestServiceServer(gRPCServer, handler)

	// Register with HTTP gateway
	ctx := context.Background()
	if err := test.RegisterTestServiceHandler(ctx, mux, conn); err != nil {
		return status.Errorf(codes.Internal, "failed to register test service handler: %v", err)
	}

	return nil
}
