package testutils

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// TestDB holds a test database connection
type TestDB struct {
	DB     *sql.DB
	Mock   sqlmock.Sqlmock
	Path   string
	IsMock bool
}

// NewTestDB creates a new test database (either mock or real SQLite)
func NewTestDB(t *testing.T, useMock bool) *TestDB {
	if useMock {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("failed to create mock database: %v", err)
		}
		return &TestDB{
			DB:     db,
			Mock:   mock,
			IsMock: true,
		}
	}

	// Create temporary SQLite database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Run schema setup
	if err := setupTestSchema(db); err != nil {
		t.Fatalf("failed to setup test schema: %v", err)
	}

	return &TestDB{
		DB:     db,
		Path:   dbPath,
		IsMock: false,
	}
}

// Close closes the test database
func (tdb *TestDB) Close() error {
	if tdb.DB != nil {
		return tdb.DB.Close()
	}
	return nil
}

// setupTestSchema creates the database schema for testing
func setupTestSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		location TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS audio_recordings (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		file_path TEXT NOT NULL,
		duration INTEGER,
		file_size INTEGER,
		sample_rate INTEGER,
		channels INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions (id)
	);

	CREATE TABLE IF NOT EXISTS sls_interactions (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		word TEXT NOT NULL,
		confidence REAL,
		frequency REAL,
		FOREIGN KEY (session_id) REFERENCES sessions (id)
	);

	CREATE TABLE IF NOT EXISTS vox_radar_detections (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		word TEXT NOT NULL,
		confidence REAL,
		FOREIGN KEY (session_id) REFERENCES sessions (id)
	);

	CREATE TABLE IF NOT EXISTS session_notes (
		id TEXT PRIMARY KEY,
		session_id TEXT NOT NULL,
		note TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES sessions (id)
	);
	`

	_, err := db.Exec(schema)
	return err
}

// AssertNoError asserts that an error is nil
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	assert.NoError(t, err, msgAndArgs...)
}

// AssertError asserts that an error is not nil
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	assert.Error(t, err, msgAndArgs...)
}

// CreateTempFile creates a temporary file with given content
func CreateTempFile(t *testing.T, content string, prefix string) string {
	tmpFile, err := os.CreateTemp("", prefix)
	AssertNoError(t, err)

	if content != "" {
		_, err = tmpFile.WriteString(content)
		AssertNoError(t, err)
	}

	err = tmpFile.Close()
	AssertNoError(t, err)

	return tmpFile.Name()
}

// CleanupFile removes a file (used in defer statements)
func CleanupFile(path string) {
	if path != "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to cleanup file %s: %v", path, err)
		}
	}
}

// GetFixturePath returns the path to a test fixture file
func GetFixturePath(filename string) string {
	return filepath.Join("testutils", "fixtures", filename)
}

// AssertMockExpectations asserts all mock expectations were met
func AssertMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	assert.NoError(t, mock.ExpectationsWereMet())
}
