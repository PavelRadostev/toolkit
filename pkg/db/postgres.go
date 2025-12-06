package db

import (
	"context"
	"fmt"
	"time"

	"github.com/PavelRadostev/toolkit/pkg/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool for convenience
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a new PostgreSQL connection pool from config
func NewPool(ctx context.Context, cfg *config.Config) (*Pool, error) {
	dsn := buildDSN(*cfg)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	// Set pool configuration
	if cfg.Postgres.MaxConns > 0 {
		poolConfig.MaxConns = int32(cfg.Postgres.MaxConns)
	}
	if cfg.Postgres.MinConns > 0 {
		poolConfig.MinConns = int32(cfg.Postgres.MinConns)
	}

	// Set connection timeout
	poolConfig.ConnConfig.ConnectTimeout = 30 * time.Second

	// Set schema search_path if specified
	if cfg.Postgres.Schema != "" {
		originalAfterConnect := poolConfig.AfterConnect
		poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			// Execute original AfterConnect hook if it exists
			if originalAfterConnect != nil {
				if err := originalAfterConnect(ctx, conn); err != nil {
					return err
				}
			}
			// Set search_path to the specified schema
			_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", pgx.Identifier{cfg.Postgres.Schema}.Sanitize()))
			return err
		}
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Close closes the connection pool
func (p *Pool) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

// buildDSN builds PostgreSQL connection string from config
func buildDSN(cfg config.Config) string {
	pg := cfg.Postgres
	sslmode := pg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	port := pg.Port
	if port == 0 {
		port = 5432
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pg.Host, port, pg.User, pg.Password, pg.DBName, sslmode,
	)
}
