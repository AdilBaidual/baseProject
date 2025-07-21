package store

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Pong implements TestRepository.Pong
func (s *Store) Pong(ctx context.Context) (string, error) {
	traceID := getTraceIDFromContext(ctx)
	s.logger.Debug("Pong method called",
		zap.String("trace_id", traceID),
		zap.String("operation", "pong"),
	)

	// Here you could add actual database logic if needed
	// For example: fetching from database, checking health, etc.

	s.logger.Debug("Pong operation completed successfully",
		zap.String("trace_id", traceID),
		zap.String("result", "pong"),
	)

	return "pong", nil
}

// getTraceIDFromContext extracts trace ID from OpenTelemetry context
func getTraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return "unknown"
	}

	// Extract trace ID from OpenTelemetry span context
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		return spanContext.TraceID().String()
	}

	return "no-trace"
}
