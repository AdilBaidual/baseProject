package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-required:"true"`
	MaxConns int32  `yaml:"max_conns"`
	MinConns int32  `yaml:"min_conns"`
}

type Storage struct {
	DB  *pgxpool.Pool
	cfg Config
}

func NewStorage(cfg Config) *Storage {
	return &Storage{cfg: cfg}
}

func (s *Storage) Connect(ctx context.Context) error {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.cfg.Host,
		s.cfg.Port,
		s.cfg.User,
		s.cfg.Password,
		s.cfg.DBName,
		s.cfg.SSLMode,
	)

	pgxConf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}

	pgxConf.MaxConns = s.cfg.MaxConns
	pgxConf.MinConns = s.cfg.MinConns

	db, err := pgxpool.NewWithConfig(ctx, pgxConf)
	if err != nil {
		return fmt.Errorf("error creating new pgx pool: %w", err)
	}

	err = db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("error connecting pgx pool: %w", err)
	}

	s.DB = db

	return nil
}

func (s *Storage) Close() {
	s.DB.Close()
}
