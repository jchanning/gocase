package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service holds the database connection pool.
type Service struct {
	pool *pgxpool.Pool
}

// NewService initializes a new database service with connection pooling.
// It verifies the connection by pinging the database.
func NewService(connString string) (*Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create the connection pool
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Service{
		pool: pool,
	}, nil
}

// Close closes the database connection pool.
func (s *Service) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Pool returns the underlying connection pool for use in queries.
func (s *Service) Pool() *pgxpool.Pool {
	return s.pool
}
