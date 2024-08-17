package config

import (
	"fmt"
	"github.com/AdilBaidual/baseProject/pkg/grpcserver"
	"github.com/AdilBaidual/baseProject/pkg/httpserver"
	"github.com/AdilBaidual/baseProject/pkg/jaeger"
	"github.com/AdilBaidual/baseProject/pkg/storage/postgres"
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
