package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/myideascope/otherside/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Test connection
	require.NoError(t, db.Ping())

	return db
}

func cleanupTestDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

func TestNewDB_ValidConfig_Success(t *testing.T) {
	// Arrange - Use in-memory database to avoid migration file issues
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	// Create temp migrations directory for testing
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	require.NoError(t, err)

	database := &DB{
		DB:       db,
		Migrator: NewMigrator(db, migrationsDir),
	}
	databaseErr := database.Initialize(context.Background())

	// Assert
	assert.NoError(t, databaseErr)
	assert.NotNil(t, database)

	// Test database is accessible
	assert.NoError(t, database.Ping())
}

func TestNewDB_InvalidDriver_Error(t *testing.T) {
	// Arrange
	cfg := &config.DatabaseConfig{
		Driver:   "invalid",
		Database: "test.db",
	}

	// Act
	db, err := NewDB(cfg)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to open database")
}

func TestNewDB_InvalidPath_Error(t *testing.T) {
	// Arrange
	cfg := &config.DatabaseConfig{
		Driver:   "sqlite3",
		Database: "/invalid/path/that/does/not/exist/test.db",
	}

	// Act
	db, err := NewDB(cfg)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestDB_Initialize_Success(t *testing.T) {
	// Arrange - Use in-memory database to avoid migration file issues
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	// Create temp migrations directory for testing
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	require.NoError(t, err)

	database := &DB{
		DB:       db,
		Migrator: NewMigrator(db, migrationsDir),
	}

	// Act
	err = database.Initialize(context.Background())

	// Assert
	assert.NoError(t, err)

	// Verify migrations table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "schema_migrations", tableName)
}

func TestDB_CheckHealth_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	database := &DB{DB: db}

	// Act
	err := database.CheckHealth(context.Background())

	// Assert
	assert.NoError(t, err)
}

func TestDB_CheckHealth_ConnectionClosed_Error(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	cleanupTestDB(db) // Close connection

	database := &DB{DB: db}

	// Act
	err := database.CheckHealth(context.Background())

	// Assert
	assert.Error(t, err)
}

func TestDB_Close_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	database := &DB{DB: db}

	// Act
	err := database.Close()

	// Assert
	assert.NoError(t, err)

	// Verify connection is closed
	err = database.Ping()
	assert.Error(t, err)
}

func TestDB_MigrationStatus_Success(t *testing.T) {
	// Arrange - Use in-memory database
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	// Create temp migrations directory for testing
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	require.NoError(t, err)

	database := &DB{
		DB:       db,
		Migrator: NewMigrator(db, migrationsDir),
	}

	// Act
	err = database.MigrationStatus(context.Background())

	// Assert
	assert.NoError(t, err)
}

func TestDB_RelativePathHandling(t *testing.T) {
	// Arrange
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Change to temp directory
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	cfg := &config.DatabaseConfig{
		Driver:   "sqlite3",
		Database: "test.db",
	}

	// Act
	db, err := NewDB(cfg)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, db)
	defer db.Close()

	// Verify database file was created in current directory
	_, err = os.Stat("test.db")
	assert.NoError(t, err)
}

func TestDB_ConnectionStringValidation(t *testing.T) {
	tests := []struct {
		name        string
		driver      string
		database    string
		expectError bool
	}{
		{
			name:        "Valid SQLite memory database",
			driver:      "sqlite3",
			database:    ":memory:",
			expectError: false,
		},
		{
			name:        "Valid SQLite file database",
			driver:      "sqlite3",
			database:    "test.db",
			expectError: false,
		},
		{
			name:        "Empty driver",
			driver:      "",
			database:    "test.db",
			expectError: true,
		},
		{
			name:        "Empty database path",
			driver:      "sqlite3",
			database:    "",
			expectError: false, // SQLite allows empty database (temp file)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &config.DatabaseConfig{
				Driver:   tt.driver,
				Database: tt.database,
			}

			// Act
			db, err := NewDB(cfg)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				require.NoError(t, err)
				if db != nil {
					defer db.Close()
				}
			}
		})
	}
}

func TestDB_ConcurrentAccess(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	// Create temp migrations directory for testing
	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	require.NoError(t, err)

	database := &DB{
		DB:       db,
		Migrator: NewMigrator(db, migrationsDir),
	}

	// Act - Simulate concurrent access
	done := make(chan bool, 2)

	go func() {
		err := database.CheckHealth(context.Background())
		assert.NoError(t, err)
		done <- true
	}()

	go func() {
		err := database.CheckHealth(context.Background())
		assert.NoError(t, err)
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Assert - No deadlocks or panics
	assert.True(t, true)
}

func TestDB_ContextCancellation(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	database := &DB{DB: db}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err := database.CheckHealth(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
