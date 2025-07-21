package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store implements all repository interfaces
type Store struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewStore creates a new Store instance that implements all repository interfaces
func NewStore(db *pgxpool.Pool, logger *zap.Logger) *Store {
	return &Store{
		db:     db,
		logger: logger,
	}
}

// Ensure Store implements TestRepository interface
var _ TestRepository = (*Store)(nil)

// GetDB returns the database connection pool
func (s *Store) GetDB() *pgxpool.Pool {
	return s.db
}

// Close closes the database connection
func (s *Store) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

// WithTx executes a function within a database transaction
func (s *Store) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.logger.Error("failed to rollback transaction after panic",
					zap.Error(rbErr),
					zap.Any("panic", p),
				)
			}
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.logger.Error("failed to rollback transaction", zap.Error(rbErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				s.logger.Error("failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	return fn(tx)
}
