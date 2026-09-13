package database

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/init_schema.up.sql
var initSchemaSQL string

//go:embed sql/seed.sql
var seedSQL string

type DB struct {
	Pool *pgxpool.Pool
}

// Connect initializes a PostgreSQL connection pool and verifies connectivity.
func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing database connection string: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping the database
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("[DATABASE_CONNECTED] Successfully connected to PostgreSQL connection pool")
	return &DB{Pool: pool}, nil
}

// RunMigrations executes initial database migration schema.
func (db *DB) RunMigrations(ctx context.Context) error {
	log.Println("[DATABASE_MIGRATION] Running database schema migrations...")
	_, err := db.Pool.Exec(ctx, initSchemaSQL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("[DATABASE_MIGRATION] Database schema migrations completed successfully")
	return nil
}

// SeedInitialData checks if media/windows are empty and populates initial sample data.
func (db *DB) SeedInitialData(ctx context.Context, force bool) error {
	if !force {
		var count int
		err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM windows").Scan(&count)
		if err == nil && count > 0 {
			log.Printf("[DATABASE_SEED] Database already contains %d windows, skipping auto-seed", count)
			return nil
		}
	}

	log.Println("[DATABASE_SEED] Seeding database with initial windows, media, and playlists...")
	_, err := db.Pool.Exec(ctx, seedSQL)
	if err != nil {
		return fmt.Errorf("failed to execute seed data: %w", err)
	}
	log.Println("[DATABASE_SEED] Initial seed data loaded successfully")
	return nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("[DATABASE_DISCONNECTED] PostgreSQL connection pool closed")
	}
}
