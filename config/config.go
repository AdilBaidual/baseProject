package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"

	"github.com/AdilBaidual/baseProject/pkg/grpcserver"
	"github.com/AdilBaidual/baseProject/pkg/httpserver"
	"github.com/AdilBaidual/baseProject/pkg/jaeger"
	"github.com/AdilBaidual/baseProject/pkg/storage/postgres"
)

const (
	defaultConfigPath = "./config/config.yaml"
	configPathEnvKey  = "CONFIG_PATH"
)

type Config struct {
	Postgres   postgres.Config   `yaml:"postgres" validate:"required"`
	Jaeger     jaeger.Config     `yaml:"jaeger" validate:"required"`
	GRPCServer grpcserver.Config `yaml:"grpc_server" validate:"required"`
	HTTPServer httpserver.Config `yaml:"http_server" validate:"required"`

	// Application-level configuration
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	Name        string `yaml:"name" env:"APP_NAME" env-default:"baseProject"`
	Version     string `yaml:"version" env:"APP_VERSION" env-default:"1.0.0"`
	Environment string `yaml:"environment" env:"APP_ENV" env-default:"development"`
	Debug       bool   `yaml:"debug" env:"DEBUG" env-default:"false"`
}

type Dependencies struct {
	Logger *zap.Logger
}

func NewConfig(deps Dependencies) (*Config, error) {
	configPath := getConfigPath()

	if deps.Logger != nil {
		deps.Logger.Info("Loading configuration", zap.String("path", configPath))
	}

	if err := validateConfigFile(configPath); err != nil {
		return nil, fmt.Errorf("config file validation failed: %w", err)
	}

	var cfg Config

	// Read from file first
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Override with environment variables
	if err := cleanenv.UpdateEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to update config from environment: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	if deps.Logger != nil {
		deps.Logger.Info("Configuration loaded successfully",
			zap.String("app.name", cfg.App.Name),
			zap.String("app.version", cfg.App.Version),
			zap.String("app.environment", cfg.App.Environment),
		)
	}

	return &cfg, nil
}

func getConfigPath() string {
	if envPath := os.Getenv(configPathEnvKey); envPath != "" {
		return envPath
	}
	return defaultConfigPath
}

func validateConfigFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file does not exist: %s", absPath)
		}
		return fmt.Errorf("failed to access config file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("config path is a directory, not a file: %s", absPath)
	}

	return nil
}

func validateConfig(cfg *Config) error {
	// Basic validation
	if cfg.App.Name == "" {
		return fmt.Errorf("app.name cannot be empty")
	}

	// Validate ports are not the same
	if cfg.GRPCServer.Port == cfg.HTTPServer.Port {
		return fmt.Errorf("gRPC and HTTP servers cannot use the same port (%d)", cfg.GRPCServer.Port)
	}

	// Validate ports are in valid range
	if cfg.GRPCServer.Port < 1024 || cfg.GRPCServer.Port > 65535 {
		return fmt.Errorf("gRPC server port must be between 1024 and 65535, got %d", cfg.GRPCServer.Port)
	}

	if cfg.HTTPServer.Port < 1024 || cfg.HTTPServer.Port > 65535 {
		return fmt.Errorf("HTTP server port must be between 1024 and 65535, got %d", cfg.HTTPServer.Port)
	}

	return nil
}

// IsProduction returns true if the application is running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if the application is running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsDebug returns true if debug mode is enabled
func (c *Config) IsDebug() bool {
	return c.App.Debug
}
