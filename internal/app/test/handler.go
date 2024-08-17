package test

import (
	"context"
	"github.com/AdilBaidual/baseProject/internal/pb/baseProject/test"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type testService interface {
	Pong() string
}

type Handler struct {
	test.TestServiceServer

	testService testService
}

func NewHandler(testService testService) *Handler {
	return &Handler{testService: testService}
}

func Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn, gRPCServer *grpc.Server, handler *Handler) error {
	test.RegisterTestServiceServer(gRPCServer, handler)
	err := test.RegisterTestServiceHandler(ctx, mux, conn)
	if err != nil {
		return err
	}
	return nil
}
