package app

import (
	"context"
	"fmt"
	"github.com/AdilBaidual/baseProject/config"
	testhandler "github.com/AdilBaidual/baseProject/internal/app/test"
	"github.com/AdilBaidual/baseProject/internal/interceptor"
	"github.com/AdilBaidual/baseProject/internal/service"
	"github.com/AdilBaidual/baseProject/internal/store"
	"github.com/AdilBaidual/baseProject/pkg/grpcserver"
	"github.com/AdilBaidual/baseProject/pkg/httpserver"
	"github.com/AdilBaidual/baseProject/pkg/jaeger"
	"github.com/AdilBaidual/baseProject/pkg/storage/postgres"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		ServiceModule(),
		JaegerModule(),
		HandlerModule(),
		DeliveryModule(),
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
	return fx.Module("postgres",
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
			store.NewStore,
		),
	)
}

func ServiceModule() fx.Option {
	return fx.Module("service",
		fx.Provide(
			service.NewServiceContainer,
		),
	)
}

func HandlerModule() fx.Option {
	return fx.Module("handler",
		fx.Provide(
			func(sc *service.ServiceContainer) *testhandler.Handler {
				return testhandler.NewHandler(sc.GetTestService())
			},
		),
	)
}

func DeliveryModule() fx.Option {
	return fx.Module("delivery",
		fx.Provide(
			func(cfg *config.Config) (grpcserver.Config, httpserver.Config) {
				return cfg.GRPCServer, cfg.HTTPServer
			},
			func() context.Context {
				return context.Background()
			},
			interceptor.NewInterceptor,
			func(ic *interceptor.Interceptor) []grpc.ServerOption {
				return []grpc.ServerOption{
					grpc.UnaryInterceptor(ic.LoggingInterceptor()),
					grpc.StatsHandler(otelgrpc.NewServerHandler()),
				}
			},
			runtime.NewServeMux,
			func(mux *runtime.ServeMux) http.Handler {
				return mux
			},
			func(cfg grpcserver.Config) (*grpc.ClientConn, error) {
				return grpc.NewClient(
					net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
			},

			grpcserver.NewServer,
			httpserver.NewServer,
			grpcserver.GetGrpcServer,
		),
		fx.Invoke(
			testhandler.Register,
			func(lc fx.Lifecycle, srv *grpcserver.Server, cfg grpcserver.Config, logger *zap.Logger, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							logger.Info(fmt.Sprintf("starting GRPC server {%s}", net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))))
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
			},
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
			},
		),
	)
}

func CheckInitializedModules() fx.Option {
	return fx.Module("check modules",
		fx.Invoke(
			func(cfg *config.Config) {},
			func(logger *zap.Logger) {},
			func(storage *postgres.Storage) {},
			func(store *store.Store) {},
			func(test *service.ServiceContainer) {},
			func(test *testhandler.Handler) {},
			//func(tracer *sdktrace.TracerProvider) {},
			func(srv *grpcserver.Server) {},
			func(srv *httpserver.Server) {},
		),
	)
}
