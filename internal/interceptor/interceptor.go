package interceptor

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type Interceptor struct {
	logger *zap.Logger
}

func NewInterceptor(logger *zap.Logger) *Interceptor {
	return &Interceptor{
		logger: logger,
	}
}

func (ic *Interceptor) LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		spanContext := trace.SpanContextFromContext(ctx)
		requestLogger := ic.logger.With(zap.String("request_id", spanContext.TraceID().String()))
		ctx = context.WithValue(ctx, "logger", requestLogger)

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		logInfos := []zap.Field{zap.String("method", info.FullMethod), zap.String("processing time", duration.String())}
		if err != nil {
			logInfos = append(logInfos, zap.String("errors", err.Error()))
		}

		requestLogger.Info("Request info", logInfos...)

		return resp, err
	}
}
