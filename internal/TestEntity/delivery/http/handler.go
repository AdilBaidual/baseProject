package http

import (
	"fmt"
	"github.com/AdilBaidual/baseProject/constant"
	"github.com/AdilBaidual/baseProject/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	engine *gin.Engine
	mw     *middleware.Middleware
}

func NewHandler(mw *middleware.Middleware) *Handler {
	return &Handler{
		mw: mw,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(constant.Mode)
	router := gin.New()
	h.engine = router

	router.Use(otelgin.Middleware(constant.ServiceName))

	api := router.Group("/api", h.mw.LoggingMiddleware())

	api.GET("/ping", func(ctx *gin.Context) {
		tracer := otel.Tracer(constant.ServiceName)
		_, span := tracer.Start(ctx.Request.Context(), "pong")
		defer span.End()

		logger, ok := ctx.Value("logger").(*zap.Logger)
		if !ok {
			fmt.Println("logger not found")
		} else {
			logger.Info("logger found!")
		}

		ctx.JSON(http.StatusOK, map[string]string{
			"message": "pong",
		})
	})

	return h.engine
}
