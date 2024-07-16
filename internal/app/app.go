package app

import (
	"Service/config"
	grpchandler "Service/internal/TestEntity/delivery/gprc"
	httphandler "Service/internal/TestEntity/delivery/http"
	"Service/internal/TestEntity/usecase"
	"Service/internal/interceptor"
	"Service/internal/middleware"
	"Service/pkg/grpcserver"
	"Service/pkg/httpserver"
	"Service/pkg/jaeger"
	"Service/pkg/storage/postgres"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strconv"
)

func NewApp() fx.Option {
	return fx.Options(
		ConfigModule(),
		LoggerModule(),
		PostgresModule(),
		RepositoryModule(),
		UseCaseModule(),
		JaegerModule(),
		HTTPModule(),
		GRPCModule(),
		CheckInitializedModules(),
	)
}

func ConfigModule() fx.Option {
	return fx.Module("config",
		fx.Provide(
			config.NewConfig,
		),
	)
}

func LoggerModule() fx.Option {
	return fx.Module("logger",
		fx.Provide(
			func() *zap.Logger {
				encoderCfg := zap.NewProductionEncoderConfig()
				encoderCfg.TimeKey = "timestamp"
				encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

				cfg := zap.Config{
					Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
					Development:       false,
					DisableCaller:     false,
					DisableStacktrace: false,
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

				return zap.Must(cfg.Build())
			},
		),
	)
}

func PostgresModule() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			func(cfg *config.Config) postgres.Config {
				return cfg.Postgres
			},
			postgres.NewStorage,
		),
		fx.Invoke(
			func(storage *postgres.Storage) error {
				return storage.Connect(context.TODO())
			},
			func(lc fx.Lifecycle, storage *postgres.Storage, logger *zap.Logger, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						storage.Close()
						return nil
					},
				})
			},
		),
	)
}

func JaegerModule() fx.Option {
	return fx.Module("jaeger",
		fx.Provide(
			func(cfg *config.Config) jaeger.Config {
				return cfg.Jaeger
			},
			jaeger.InitJaeger,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, tracer *sdktrace.TracerProvider, cfg jaeger.Config, logger *zap.Logger, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return nil
					},
					OnStop: func(ctx context.Context) error {
						return nil
					},
				})
			},
		),
	)
}

func RepositoryModule() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			func(storage *postgres.Storage) *pgxpool.Pool {
				return storage.DB
			},
		),
	)
}

func UseCaseModule() fx.Option {
	return fx.Module("usecase",
		fx.Provide(
			usecase.NewTestUseCase,
		),
	)
}

func HTTPModule() fx.Option {
	return fx.Module("http server",
		fx.Provide(
			func(cfg *config.Config) httpserver.Config {
				return cfg.HTTPServer
			},
			middleware.NewMiddleware,
			httphandler.NewHandler,
			fx.Annotate(
				func(h *httphandler.Handler) *gin.Engine {
					return h.InitRoutes()
				},
				fx.As(new(http.Handler)),
			),
			httpserver.NewServer,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, srv *httpserver.Server, cfg httpserver.Config, logger *zap.Logger, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							logger.Info(fmt.Sprintf("starting HTTP server {%s}", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))))
							if err := srv.Start(); err != nil {
								logger.Error("error starting HTTP server",
									zap.Error(err),
									zap.String("address", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))),
								)
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						if err := srv.Stop(ctx); err != nil {
							logger.Error("error stopping HTTP server",
								zap.Error(err),
								zap.String("address", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))),
							)
						}
						return nil
					},
				})
			}),
	)
}

func GRPCModule() fx.Option {
	return fx.Module("grpc server",
		fx.Provide(
			func(cfg *config.Config) grpcserver.Config {
				return cfg.GRPCServer
			},
			interceptor.NewInterceptor,
			func(ic *interceptor.Interceptor) []grpc.ServerOption {
				return []grpc.ServerOption{
					grpc.UnaryInterceptor(ic.LoggingInterceptor()),
					grpc.StatsHandler(otelgrpc.NewServerHandler()),
				}
			},
			grpcserver.NewServer,
		),
		fx.Invoke(
			func(srv *grpcserver.Server) {
				grpchandler.Register(srv.Srv)
			},
			func(lc fx.Lifecycle, srv *grpcserver.Server, cfg grpcserver.Config, logger *zap.Logger, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							logger.Info(fmt.Sprintf("starting grpc server {%s}", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))))
							if err := srv.Start(); err != nil {
								logger.Error("error starting GRPC server",
									zap.Error(err),
									zap.String("address", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))),
								)
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						srv.Stop()
						return nil
					},
				})
			}),
	)
}

func CheckInitializedModules() fx.Option {
	return fx.Module("check modules",
		fx.Invoke(
			func(cfg *config.Config) {},
			func(logger *zap.Logger) {},
			func(storage *postgres.Storage) {},
			func(tracer *sdktrace.TracerProvider) {},
			func(srv *grpcserver.Server) {},
			func(srv *httpserver.Server) {},
		),
	)
}
