package config

import (
	"Service/pkg/grpcserver"
	"Service/pkg/httpserver"
	"Service/pkg/jaeger"
	"Service/pkg/storage/postgres"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

const configPath string = "./config/config.yaml"

type Config struct {
	Postgres   postgres.Config   `yaml:"postgres"`
	Jaeger     jaeger.Config     `yaml:"jaeger"`
	GRPCServer grpcserver.Config `yaml:"grpc_server"`
	HTTPServer httpserver.Config `yaml:"http_server"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		return nil, fmt.Errorf("NewConfig - cleanenv.ReadConfig - %w", err)
	}

	err = cleanenv.UpdateEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("NewConfig - cleanenv.UpdateEnv - %w", err)
	}

	return &cfg, nil
}
