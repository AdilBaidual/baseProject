package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"
)

type Middleware struct {
	logger *zap.Logger
}

func NewMiddleware(logger *zap.Logger) *Middleware {
	return &Middleware{
		logger: logger,
	}
}

func (mw *Middleware) LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		start := time.Now()

		spanContext := trace.SpanContextFromContext(ctx.Request.Context())
		requestLogger := mw.logger.With(zap.String("request_id", spanContext.TraceID().String()))
		ctx.Set("logger", requestLogger)

		ctx.Next()

		duration := time.Since(start)

		logInfos := []zap.Field{zap.String("method", ctx.Request.URL.Path), zap.String("processing time", duration.String())}
		if ctx.Errors != nil {
			logInfos = append(logInfos, zap.String("errors", ctx.Errors.String()))
		}

		requestLogger.Info("Request info", logInfos...)
	}
}
