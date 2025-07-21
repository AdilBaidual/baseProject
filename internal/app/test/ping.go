package test

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/AdilBaidual/baseProject/constant"
	"github.com/AdilBaidual/baseProject/internal/interceptor"
	"github.com/AdilBaidual/baseProject/internal/pb/baseProject/test"
)

// Ping handles the ping gRPC request
func (h *Handler) Ping(ctx context.Context, _ *emptypb.Empty) (*test.PingResponse, error) {
	// Create tracing span
	tracer := otel.Tracer(constant.ServiceName)
	ctx, span := tracer.Start(ctx, "Handler.Ping")
	defer span.End()

	// Get logger from context or use handler's logger
	logger := interceptor.GetLoggerFromContext(ctx, h.logger)
	logger.Info("Processing ping request")

	// Call service layer
	message, err := h.service.Pong(ctx)
	if err != nil {
		logger.Error("Failed to process ping request", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to process ping request: %v", err)
	}

	logger.Info("Ping request processed successfully", zap.String("message", message))

	return &test.PingResponse{Message: message}, nil
}
