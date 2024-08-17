package grpc

import (
	"context"
	"fmt"
	"github.com/AdilBaidual/baseProject/constant"
	"github.com/AdilBaidual/baseProject/internal/pb/baseProject/test"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	test.TestServiceServer
}

func Register(gRPCServer *grpc.Server) {
	test.RegisterTestServiceServer(gRPCServer, &Handler{})
}

func (h *Handler) Ping(ctx context.Context, _ *emptypb.Empty) (*test.PingResponse, error) {
	tracer := otel.Tracer(constant.ServiceName)
	_, span := tracer.Start(ctx, "pong")
	defer span.End()

	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		fmt.Println("logger not found")
	} else {
		logger.Info("logger found!")
	}

	return &test.PingResponse{Message: "pong"}, nil
}
