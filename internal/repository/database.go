package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/myideascope/otherside/internal/config"
)

// DB holds the database connection and migration system
type DB struct {
	*sql.DB
	Migrator *Migrator
}

// NewDB creates a new database connection with migration support
func NewDB(cfg *config.DatabaseConfig) (*DB, error) {
	// For SQLite, construct the database path
	dbPath := cfg.Database
	if cfg.Driver == "sqlite3" && !filepath.IsAbs(dbPath) {
		// If relative path, make it relative to current working directory
		dbPath = filepath.Join(".", dbPath)
	}

	// Open database connection
	db, err := sql.Open(cfg.Driver, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize migrator
	migrationsPath := filepath.Join("internal", "repository", "migrations")
	migrator := NewMigrator(db, migrationsPath)

	database := &DB{
		DB:       db,
		Migrator: migrator,
	}

	// Initialize migrations system
	if err := database.Initialize(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// Initialize sets up the database and runs migrations
func (db *DB) Initialize(ctx context.Context) error {
	log.Println("Initializing database...")

	// Create migrations table
	if err := db.Migrator.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Run pending migrations
	if err := db.Migrator.Up(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get current version
	version, err := db.Migrator.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if version != "" {
		log.Printf("Database initialized successfully. Current version: %s", version)
	} else {
		log.Println("Database initialized successfully. No migrations applied.")
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	log.Println("Closing database connection...")
	return db.DB.Close()
}

// CheckHealth verifies the database is accessible
func (db *DB) CheckHealth(ctx context.Context) error {
	return db.PingContext(ctx)
}

// MigrationStatus shows the current migration status
func (db *DB) MigrationStatus(ctx context.Context) error {
	return db.Migrator.Status(ctx)
}
