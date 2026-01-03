package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMigrationTestDB(t *testing.T) *sql.DB {
	db := setupTestDB(t)
	return db
}

func setupMigrationTestFiles(t *testing.T, tempDir string) string {
	migrationsDir := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsDir, 0755)
	require.NoError(t, err)

	// Create test migration files
	migrations := map[string]string{
		"001_initial_schema.sql": `
			CREATE TABLE test_table (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL
			);
		`,
		"002_add_indexes.sql": `
			CREATE INDEX idx_test_table_name ON test_table(name);
		`,
		"003_add_another_table.sql": `
			CREATE TABLE another_table (
				id TEXT PRIMARY KEY,
				description TEXT
			);
		`,
	}

	for filename, content := range migrations {
		filePath := filepath.Join(migrationsDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	return migrationsDir
}

func TestNewMigrator_ValidInput_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := "/tmp/migrations"

	// Act
	migrator := NewMigrator(db, migrationsPath)

	// Assert
	assert.NotNil(t, migrator)
	assert.Equal(t, db, migrator.db)
	assert.Equal(t, migrationsPath, migrator.migrationsPath)
}

func TestMigrator_Initialize_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	// Act
	err := migrator.Initialize(context.Background())

	// Assert
	assert.NoError(t, err)

	// Verify migrations table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "schema_migrations", tableName)
}

func TestMigrator_LoadMigrations_ValidDirectory_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	// Act
	err := migrator.LoadMigrations()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, migrator.migrations, 3)

	// Verify migrations are sorted by version
	assert.Equal(t, "001", migrator.migrations[0].Version)
	assert.Equal(t, "002", migrator.migrations[1].Version)
	assert.Equal(t, "003", migrator.migrations[2].Version)
}

func TestMigrator_LoadMigrations_InvalidDirectory_Error(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrator := NewMigrator(db, "/non/existent/directory")

	// Act
	err := migrator.LoadMigrations()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read migrations directory")
}

func TestMigrator_LoadMigrations_IgnoreNonSQLFiles_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())

	// Add non-SQL files
	nonSQLFiles := []string{"README.md", "backup.txt", "test.sql.bak"}
	for _, filename := range nonSQLFiles {
		filePath := filepath.Join(migrationsPath, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	migrator := NewMigrator(db, migrationsPath)

	// Act
	err := migrator.LoadMigrations()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, migrator.migrations, 3) // Should only load SQL files
}

func TestMigrator_GetAppliedMigrations_EmptyDatabase_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Act
	applied, err := migrator.GetAppliedMigrations(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, applied)
}

func TestMigrator_GetAppliedMigrations_WithMigrations_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Manually insert a migration
	_, err = db.Exec("INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, datetime('now'))",
		"001", "initial_schema")
	require.NoError(t, err)

	// Act
	applied, err := migrator.GetAppliedMigrations(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Len(t, applied, 1)
	assert.True(t, applied["001"])
	assert.False(t, applied["002"])
	assert.False(t, applied["003"])
}

func TestMigrator_Up_AllPendingMigrations_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Act
	err = migrator.Up(context.Background())

	// Assert
	assert.NoError(t, err)

	// Verify all migrations were applied
	applied, err := migrator.GetAppliedMigrations(context.Background())
	assert.NoError(t, err)
	assert.Len(t, applied, 3)
	assert.True(t, applied["001"])
	assert.True(t, applied["002"])
	assert.True(t, applied["003"])

	// Verify tables were created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "test_table", tableName)

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='another_table'").Scan(&tableName)
	assert.NoError(t, err)
	assert.Equal(t, "another_table", tableName)
}

func TestMigrator_Up_PartialMigrations_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Apply first migration manually with schema
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, datetime('now'))",
		"001", "initial_schema")
	require.NoError(t, err)

	// Act
	err = migrator.Up(context.Background())

	// Assert
	assert.NoError(t, err)

	// Verify remaining migrations were applied
	applied, err := migrator.GetAppliedMigrations(context.Background())
	assert.NoError(t, err)
	assert.Len(t, applied, 3)
	assert.True(t, applied["001"])
	assert.True(t, applied["002"])
	assert.True(t, applied["003"])
}

func TestMigrator_Up_AllMigrationsAlreadyApplied_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Apply all migrations manually
	for _, version := range []string{"001", "002", "003"} {
		_, err = db.Exec("INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, datetime('now'))",
			version, "test_migration")
		require.NoError(t, err)
	}

	// Act
	err = migrator.Up(context.Background())

	// Assert
	assert.NoError(t, err)

	// Verify no duplicate migrations
	applied, err := migrator.GetAppliedMigrations(context.Background())
	assert.NoError(t, err)
	assert.Len(t, applied, 3)
}

func TestMigrator_Up_InvalidSQLMigration_RollbackError(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	// Create migration with invalid SQL
	invalidSQL := `
		CREATE TABLE test_table (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
		INVALID SQL SYNTAX HERE;
	`

	filePath := filepath.Join(migrationsPath, "001_invalid_migration.sql")
	err = os.WriteFile(filePath, []byte(invalidSQL), 0644)
	require.NoError(t, err)

	migrator := NewMigrator(db, migrationsPath)

	err = migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Act
	err = migrator.Up(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute migration 001")

	// Verify no migrations were applied
	applied, err := migrator.GetAppliedMigrations(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, applied)
}

func TestMigrator_Status_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Apply one migration manually
	_, err = db.Exec("INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, datetime('now'))",
		"001", "initial_schema")
	require.NoError(t, err)

	// Act
	err = migrator.Status(context.Background())

	// Assert
	assert.NoError(t, err)
}

func TestMigrator_GetCurrentVersion_NoMigrations_EmptyString(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Act
	version, err := migrator.GetCurrentVersion(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, version)
}

func TestMigrator_GetCurrentVersion_WithMigrations_ReturnsLatest(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Apply all migrations
	err = migrator.Up(context.Background())
	require.NoError(t, err)

	// Act
	version, err := migrator.GetCurrentVersion(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "003", version) // Should return the latest version
}

func TestMigrator_CreateInitialMigration_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	migrator := NewMigrator(db, migrationsPath)

	schemaContent := `
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL
		);
	`

	schemaPath := filepath.Join(tempDir, "schema.sql")
	err = os.WriteFile(schemaPath, []byte(schemaContent), 0644)
	require.NoError(t, err)

	// Act
	err = migrator.CreateInitialMigration(schemaPath)

	// Assert
	assert.NoError(t, err)

	// Verify migration file was created
	migrationFile := filepath.Join(migrationsPath, "001_initial_schema.sql")
	_, err = os.Stat(migrationFile)
	assert.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(migrationFile)
	assert.NoError(t, err)
	assert.Equal(t, schemaContent, string(content))
}

func TestMigrator_MigrationFileNameParsing_Success(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	tempDir := t.TempDir()
	migrationsPath := filepath.Join(tempDir, "migrations")
	err := os.MkdirAll(migrationsPath, 0755)
	require.NoError(t, err)

	// Create migration with complex filename
	filename := "001_complex_migration_with_underscores.sql"
	content := "CREATE TABLE test (id TEXT PRIMARY KEY);"
	filePath := filepath.Join(migrationsPath, filename)
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	migrator := NewMigrator(db, migrationsPath)

	// Act
	err = migrator.LoadMigrations()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, migrator.migrations, 1)
	assert.Equal(t, "001", migrator.migrations[0].Version)
	assert.Equal(t, "complex_migration_with_underscores", migrator.migrations[0].Description)
	assert.Equal(t, content, migrator.migrations[0].SQL)
}

func TestMigrator_ContextCancellation_Error(t *testing.T) {
	// Arrange
	db := setupMigrationTestDB(t)
	defer cleanupTestDB(db)
	migrationsPath := setupMigrationTestFiles(t, t.TempDir())
	migrator := NewMigrator(db, migrationsPath)

	err := migrator.Initialize(context.Background())
	require.NoError(t, err)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err = migrator.Up(ctx)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
