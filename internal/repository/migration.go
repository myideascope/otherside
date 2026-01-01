package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Migration represents a database migration
type Migration struct {
	Version     string     `json:"version"`
	Description string     `json:"description"`
	SQL         string     `json:"sql"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
}

// Migrator handles database migrations
type Migrator struct {
	db             *sql.DB
	migrations     []Migration
	migrationsPath string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

// Initialize creates the migrations table if it doesn't exist
func (m *Migrator) Initialize(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`

	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	return nil
}

// LoadMigrations loads migration files from the migrations directory
func (m *Migrator) LoadMigrations() error {
	files, err := os.ReadDir(m.migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Extract version and description from filename
		// Format: 001_initial_schema.sql, 002_add_indexes.sql, etc.
		filename := file.Name()
		parts := strings.Split(strings.TrimSuffix(filename, ".sql"), "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		description := strings.Join(parts[1:], "_")

		// Read SQL content
		filePath := filepath.Join(m.migrationsPath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	m.migrations = migrations
	return nil
}

// GetAppliedMigrations retrieves the list of applied migrations
func (m *Migrator) GetAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	query := "SELECT version FROM schema_migrations"
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// Up runs all pending migrations
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.LoadMigrations(); err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %s: %s", migration.Version, migration.Description)

		// Begin transaction
		tx, err := m.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", migration.Version, err)
		}

		// Execute migration SQL
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		// Record migration as applied
		insertQuery := "INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)"
		now := time.Now()
		if _, err := tx.ExecContext(ctx, insertQuery, migration.Version, migration.Description, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		log.Printf("Migration %s applied successfully", migration.Version)
	}

	return nil
}

// Status shows the migration status
func (m *Migrator) Status(ctx context.Context) error {
	if err := m.LoadMigrations(); err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	applied, err := m.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	log.Println("Migration Status:")
	log.Println("=================")

	for _, migration := range m.migrations {
		status := "Pending"
		if applied[migration.Version] {
			status = "Applied"
		}
		log.Printf("%s %s: %s", migration.Version, status, migration.Description)
	}

	return nil
}

// GetCurrentVersion returns the current database version
func (m *Migrator) GetCurrentVersion(ctx context.Context) (string, error) {
	query := "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1"
	var version string
	err := m.db.QueryRowContext(ctx, query).Scan(&version)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get current version: %w", err)
	}
	return version, nil
}

// CreateInitialMigration creates the initial migration from the existing schema
func (m *Migrator) CreateInitialMigration(schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	migration := Migration{
		Version:     "001",
		Description: "initial_schema",
		SQL:         string(content),
	}

	filename := fmt.Sprintf("%s_%s.sql", migration.Version, migration.Description)
	migrationPath := filepath.Join(m.migrationsPath, filename)

	// Ensure migrations directory exists
	if err := os.WriteFile(migrationPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	log.Printf("Created initial migration: %s", filename)
	return nil
}
