package store

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (s *Store) Pong(ctx context.Context) (string, error) {
	traceID := getTraceIDFromContext(ctx)
	s.logger.Debug("Pong method called",
		zap.String("trace_id", traceID),
		zap.String("operation", "pong"),
	)

	s.logger.Debug("Pong operation completed successfully",
		zap.String("trace_id", traceID),
		zap.String("result", "pong"),
	)

	return "pong", nil
}

func getTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return "unknown"
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		return spanContext.TraceID().String()
	}

	return "no-trace"
}
