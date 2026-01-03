package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/myideascope/otherside/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSessionStateManager_Initialize_LoadsActiveAndPausedSessions_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	// Insert test sessions
	_, err = db.Exec(`
		INSERT INTO sessions (id, title, status, created_at, updated_at) VALUES
		('test-active-1', 'Active Session', 'active', '2024-01-01T10:00:00Z', '2024-01-01T10:00:00Z'),
		('test-paused-1', 'Paused Session', 'paused', '2024-01-01T09:00:00Z', '2024-01-01T09:00:00Z'),
		('test-completed-1', 'Completed Session', 'complete', '2024-01-01T08:00:00Z', '2024-01-01T08:00:00Z')
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)

	// Act
	err = sm.Initialize(context.Background())

	// Assert
	require.NoError(t, err)

	// Verify we have active and paused sessions in memory
	activeSessions := sm.GetActiveSessions()
	require.Len(t, activeSessions, 2) // active and paused sessions
}

func TestSessionStateManager_CreateSession_ValidSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()

	// Act
	err = sm.CreateSession(context.Background(), session)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusActive, session.Status)
	assert.NotEmpty(t, session.CreatedAt)
	assert.NotEmpty(t, session.UpdatedAt)

	// Verify session is in active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, session.ID, activeSessions[0].ID)
}

func TestSessionStateManager_GetSession_InMemory_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Act
	retrievedSession, err := sm.GetSession(context.Background(), session.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, session.ID, retrievedSession.ID)
	assert.Equal(t, session.Title, retrievedSession.Title)
	assert.Equal(t, session.Status, retrievedSession.Status)
}

func TestSessionStateManager_GetSession_FromDatabase_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	// Insert session directly to database
	_, err = db.Exec(`
		INSERT INTO sessions (id, title, status, created_at, updated_at) VALUES
		('test-db-session', 'Database Session', 'active', '2024-01-01T10:00:00Z', '2024-01-01T10:00:00Z')
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	// Act
	retrievedSession, err := sm.GetSession(context.Background(), "test-db-session")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "test-db-session", retrievedSession.ID)
	assert.Equal(t, "Database Session", retrievedSession.Title)
	assert.Equal(t, domain.SessionStatusActive, retrievedSession.Status)

	// Verify session was cached
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, "test-db-session", activeSessions[0].ID)
}

func TestSessionStateManager_UpdateSession_ValidSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	originalUpdatedAt := session.UpdatedAt

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Update session
	session.Title = "Updated Session Title"
	session.Notes = "Updated notes"

	// Act
	err = sm.UpdateSession(context.Background(), session)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Updated Session Title", session.Title)
	assert.Equal(t, "Updated notes", session.Notes)
	assert.True(t, session.UpdatedAt.After(originalUpdatedAt))

	// Verify session is still in active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
}

func TestSessionStateManager_PauseSession_ActiveSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Act
	err = sm.PauseSession(context.Background(), session.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusPaused, session.Status)

	// Verify session is still in active sessions (paused sessions are kept)
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
}

func TestSessionStateManager_ResumeSession_PausedSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// First pause the session
	err = sm.PauseSession(context.Background(), session.ID)
	require.NoError(t, err)

	// Act
	err = sm.ResumeSession(context.Background(), session.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusActive, session.Status)

	// Verify session is in active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
}

func TestSessionStateManager_CompleteSession_ActiveSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Act
	err = sm.CompleteSession(context.Background(), session.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusComplete, session.Status)
	assert.NotNil(t, session.EndTime)

	// Verify session was removed from active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 0)
}

func TestSessionStateManager_ArchiveSession_ActiveSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Act
	err = sm.ArchiveSession(context.Background(), session.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusArchived, session.Status)

	// Verify session was removed from active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 0)
}

func TestSessionStateManager_DeleteSession_ActiveSession_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Act
	err = sm.DeleteSession(context.Background(), session.ID)

	// Assert
	require.NoError(t, err)

	// Verify session was removed from active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 0)
}

func TestSessionStateManager_GetActiveSessionsByStatus_MixedStatuses_ReturnsFiltered(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	// Create sessions with different statuses
	activeSession := TestSession()
	activeSession.ID = "active-test"
	err = sm.CreateSession(context.Background(), activeSession)
	require.NoError(t, err)

	pausedSession := TestSession()
	pausedSession.ID = "paused-test"
	err = sm.CreateSession(context.Background(), pausedSession)
	require.NoError(t, err)
	err = sm.PauseSession(context.Background(), pausedSession.ID)
	require.NoError(t, err)

	// Act
	activeSessions := sm.GetActiveSessionsByStatus(domain.SessionStatusActive)
	pausedSessions := sm.GetActiveSessionsByStatus(domain.SessionStatusPaused)

	// Assert
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, "active-test", activeSessions[0].ID)
	assert.Equal(t, domain.SessionStatusActive, activeSessions[0].Status)

	assert.Len(t, pausedSessions, 1)
	assert.Equal(t, "paused-test", pausedSessions[0].ID)
	assert.Equal(t, domain.SessionStatusPaused, pausedSessions[0].Status)
}

func TestSessionStateManager_SaveSessionState_ActiveSessions_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Modify session to ensure it gets saved
	session.Notes = "Updated for save test"

	// Act
	err = sm.SaveSessionState(context.Background())

	// Assert
	require.NoError(t, err)

	// Verify session is still in memory
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, "Updated for save test", activeSessions[0].Notes)
}

func TestSessionStateManager_CleanupExpiredSessions_ExpiredSessions_Archived(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	// Create old session (manually set old updated_at time)
	oldSession := TestSession()
	oldSession.ID = "old-session"
	oldSession.UpdatedAt = time.Now().Add(-2 * time.Hour) // 2 hours ago
	err = sm.CreateSession(context.Background(), oldSession)
	require.NoError(t, err)

	// Create recent session
	recentSession := TestSession()
	recentSession.ID = "recent-session"
	err = sm.CreateSession(context.Background(), recentSession)
	require.NoError(t, err)

	// Act - cleanup sessions inactive for more than 1 hour
	err = sm.CleanupExpiredSessions(context.Background(), 1*time.Hour)

	// Assert
	require.NoError(t, err)

	// Verify old session was archived
	retrievedOldSession, err := sm.GetSession(context.Background(), "old-session")
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusArchived, retrievedOldSession.Status)

	// Verify recent session is still active
	retrievedRecentSession, err := sm.GetSession(context.Background(), "recent-session")
	require.NoError(t, err)
	assert.Equal(t, domain.SessionStatusActive, retrievedRecentSession.Status)

	// Verify old session was removed from active sessions
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, "recent-session", activeSessions[0].ID)
}

func TestSessionStateManager_Shutdown_SaveState_Success(t *testing.T) {
	// Arrange
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT,
			location TEXT,
			start_time TEXT,
			end_time TEXT,
			notes TEXT,
			environmental TEXT,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		)
	`)
	require.NoError(t, err)

	sm := NewSessionStateManager(db)
	err = sm.Initialize(context.Background())
	require.NoError(t, err)

	session := TestSession()
	err = sm.CreateSession(context.Background(), session)
	require.NoError(t, err)

	// Act
	err = sm.Shutdown(context.Background())

	// Assert
	require.NoError(t, err)

	// Verify session was properly saved before shutdown
	activeSessions := sm.GetActiveSessions()
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, session.ID, activeSessions[0].ID)
}
