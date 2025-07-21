package app

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/AdilBaidual/baseProject/config"
	testhandler "github.com/AdilBaidual/baseProject/internal/app/test"
	"github.com/AdilBaidual/baseProject/internal/interceptor"
	"github.com/AdilBaidual/baseProject/internal/service/testservice"
	"github.com/AdilBaidual/baseProject/internal/store"
	"github.com/AdilBaidual/baseProject/pkg/grpcserver"
	"github.com/AdilBaidual/baseProject/pkg/httpserver"
	"github.com/AdilBaidual/baseProject/pkg/jaeger"
	"github.com/AdilBaidual/baseProject/pkg/storage/postgres"
)

// NewApp creates and configures the FX application
func NewApp() fx.Option {
	return fx.Options(
		// Core modules
		ConfigModule(),
		LoggerModule(),
		DatabaseModule(),
		TracingModule(),

		// Business logic modules
		RepositoryModule(),
		ServiceModule(),
		HandlerModule(),

		// Server modules
		ServerModule(),

		// Lifecycle management
		fx.Invoke(registerLifecycleHooks),
	)
}

// ConfigModule provides application configuration
func ConfigModule() fx.Option {
	return fx.Module("config",
		fx.Provide(
			func() config.Dependencies {
				return config.Dependencies{Logger: nil} // Logger will be injected later if needed
			},
			config.NewConfig,
		),
	)
}

// LoggerModule provides structured logging
func LoggerModule() fx.Option {
	return fx.Module("logger",
		fx.Provide(
			func(cfg *config.Config) *zap.Logger {
				var level zapcore.Level
				if cfg.IsDebug() {
					level = zap.DebugLevel
				} else if cfg.IsDevelopment() {
					level = zap.InfoLevel
				} else {
					level = zap.WarnLevel
				}

				encoderCfg := zap.NewProductionEncoderConfig()
				encoderCfg.TimeKey = "timestamp"
				encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

				config := zap.Config{
					Level:             zap.NewAtomicLevelAt(level),
					Development:       cfg.IsDevelopment(),
					DisableCaller:     false,
					DisableStacktrace: cfg.IsProduction(),
					Sampling:          nil,
					Encoding:          "json",
					EncoderConfig:     encoderCfg,
					OutputPaths: []string{
						"stderr",
					},
					ErrorOutputPaths: []string{
						"stderr",
					},
				}

				return zap.Must(config.Build())
			},
		),
	)
}

// DatabaseModule provides database connectivity
func DatabaseModule() fx.Option {
	return fx.Module("database",
		fx.Provide(
			func(cfg *config.Config) postgres.Config {
				return cfg.Postgres
			},
			postgres.NewStorage,
			func(storage *postgres.Storage) *pgxpool.Pool {
				return storage.DB
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, storage *postgres.Storage, logger *zap.Logger) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						logger.Info("Connecting to database...")
						if err := storage.Connect(ctx); err != nil {
							logger.Error("Failed to connect to database", zap.Error(err))
							return err
						}
						logger.Info("Database connection established")
						return nil
					},
					OnStop: func(_ context.Context) error {
						logger.Info("Closing database connection...")
						storage.Close()
						logger.Info("Database connection closed")
						return nil
					},
				})
			},
		),
	)
}

// TracingModule provides distributed tracing
func TracingModule() fx.Option {
	return fx.Module("tracing",
		fx.Provide(
			func(cfg *config.Config) jaeger.Config {
				return cfg.Jaeger
			},
			jaeger.InitJaeger,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, tracer *sdktrace.TracerProvider, logger *zap.Logger) {
				lc.Append(fx.Hook{
					OnStart: func(_ context.Context) error {
						logger.Info("Tracing initialized")
						return nil
					},
					OnStop: func(ctx context.Context) error {
						shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
						defer cancel()

						if err := tracer.Shutdown(shutdownCtx); err != nil {
							logger.Error("Failed to shutdown tracer", zap.Error(err))
							return err
						}
						logger.Info("Tracer shutdown completed")
						return nil
					},
				})
			},
		),
	)
}

// RepositoryModule provides repository implementations
func RepositoryModule() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			store.NewStore,
			// Provide store as TestRepository interface
			func(store *store.Store) testservice.TestRepository {
				return store
			},
		),
	)
}

// ServiceModule provides business logic services
func ServiceModule() fx.Option {
	return fx.Module("service",
		fx.Provide(
			testservice.NewService,
			// Provide service as Service interface
			func(service *testservice.Service) testhandler.Service {
				return service
			},
		),
	)
}

// HandlerModule provides gRPC handlers
func HandlerModule() fx.Option {
	return fx.Module("handler",
		fx.Provide(
			testhandler.NewHandler,
		),
	)
}

// ServerModule provides HTTP and gRPC servers
func ServerModule() fx.Option {
	return fx.Module("server",
		fx.Provide(
			// Server configurations
			func(cfg *config.Config) (grpcserver.Config, httpserver.Config) {
				return cfg.GRPCServer, cfg.HTTPServer
			},

			// Interceptors and middleware
			interceptor.NewInterceptor,
			func(ic *interceptor.Interceptor) []grpc.ServerOption {
				return []grpc.ServerOption{
					grpc.UnaryInterceptor(ic.LoggingInterceptor()),
					grpc.StatsHandler(otelgrpc.NewServerHandler()),
				}
			},

			// gRPC Gateway
			runtime.NewServeMux,

			// Simple HTTP Handler
			func(mux *runtime.ServeMux) http.Handler {
				return mux
			},

			// gRPC client for gateway
			func(cfg grpcserver.Config) (*grpc.ClientConn, error) {
				return grpc.NewClient(
					net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
			},

			// Servers
			grpcserver.NewServer,
			httpserver.NewServer,
			grpcserver.GetGrpcServer,
		),
		fx.Invoke(
			// Register handlers
			testhandler.Register,
		),
	)
}

// registerLifecycleHooks registers application lifecycle hooks
func registerLifecycleHooks(
	lc fx.Lifecycle,
	grpcSrv *grpcserver.Server,
	httpSrv *httpserver.Server,
	grpcCfg grpcserver.Config,
	httpCfg httpserver.Config,
	logger *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			// Start gRPC server
			go func() {
				addr := net.JoinHostPort(grpcCfg.Host, strconv.Itoa(grpcCfg.Port))
				logger.Info("Starting gRPC server", zap.String("address", addr))
				if err := grpcSrv.Start(); err != nil {
					logger.Error("gRPC server failed", zap.Error(err))
				}
			}()

			// Start HTTP server
			go func() {
				addr := net.JoinHostPort(httpCfg.Host, strconv.Itoa(httpCfg.Port))
				logger.Info("Starting HTTP server", zap.String("address", addr))
				if err := httpSrv.Start(); err != nil && err != http.ErrServerClosed {
					logger.Error("HTTP server failed", zap.Error(err))
				}
			}()

			// Give servers a moment to start
			time.Sleep(100 * time.Millisecond)
			logger.Info("Application started successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down application...")

			// Create shutdown context with timeout
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			// Stop gRPC server
			logger.Info("Stopping gRPC server...")
			grpcSrv.Stop()

			// Stop HTTP server
			logger.Info("Stopping HTTP server...")
			if err := httpSrv.Stop(shutdownCtx); err != nil {
				logger.Error("Failed to stop HTTP server gracefully", zap.Error(err))
				return err
			}

			logger.Info("Application shutdown completed")
			return nil
		},
	})
}
