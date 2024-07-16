package grpcserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"strconv"
	"time"
)

type Config struct {
	Host              string        `env:"GRPC_SERVER_HOST" env-required:"true"`
	Port              int           `env:"GRPC_SERVER_PORT" env-required:"true"`
	MaxConnectionIdle time.Duration `yaml:"max_connection_idle"`
	MaxConnectionAge  time.Duration `yaml:"max_connection_age"`
	Time              time.Duration `yaml:"time"`
	Timeout           time.Duration `yaml:"timeout"`
}

type Server struct {
	Srv *grpc.Server
	Cfg Config
}

func NewServer(cfg Config, opts []grpc.ServerOption) *Server {
	serverOptions := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.MaxConnectionIdle,
			MaxConnectionAge:  cfg.MaxConnectionAge,
			Time:              cfg.Time,
			Timeout:           cfg.Timeout,
		}),
	}

	serverOptions = append(serverOptions, opts...)

	srv := grpc.NewServer(serverOptions...)

	return &Server{
		Srv: srv,
		Cfg: cfg,
	}
}

func (s *Server) Start() error {
	lsn, err := net.Listen("tcp", net.JoinHostPort(s.Cfg.Host, strconv.Itoa(s.Cfg.Port)))
	if err != nil {
		return err
	}

	return s.Srv.Serve(lsn)
}

func (s *Server) Stop() {
	s.Srv.GracefulStop()
}
