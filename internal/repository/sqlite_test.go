package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/myideascope/otherside/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSchema(t *testing.T, db *sql.DB) {
	schema := `
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			location_latitude REAL,
			location_longitude REAL,
			location_address TEXT,
			location_description TEXT,
			location_venue TEXT,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			notes TEXT,
			env_temperature REAL,
			env_humidity REAL,
			env_pressure REAL,
			env_emf_level REAL,
			env_light_level REAL,
			env_noise_level REAL,
			status TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE evp_recordings (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			duration REAL NOT NULL,
			timestamp DATETIME NOT NULL,
			waveform_data TEXT,
			processed_path TEXT,
			annotations TEXT,
			quality TEXT NOT NULL,
			detection_level REAL NOT NULL,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE vox_events (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			language_pack TEXT NOT NULL,
			trigger_strength REAL NOT NULL,
			frequency_data TEXT,
			user_response TEXT,
			response_delay INTEGER,
			randomizer_result TEXT,
			audio_path TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE radar_events (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			source_type TEXT NOT NULL,
			strength REAL NOT NULL,
			position_x REAL NOT NULL,
			position_y REAL NOT NULL,
			movement_trail TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE sls_detections (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			confidence REAL NOT NULL,
			duration INTEGER NOT NULL,
			skeletal_points TEXT,
			bounding_box TEXT,
			movement_pattern TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE user_interactions (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			interaction_type TEXT NOT NULL,
			response_time INTEGER,
			randomizer_result TEXT,
			audio_path TEXT,
			created_at DATETIME NOT NULL
		);

		CREATE TABLE files (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			file_type TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			mime_type TEXT NOT NULL,
			checksum TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);
	`

	_, err := db.Exec(schema)
	require.NoError(t, err)
}

func createTestSession() *domain.Session {
	now := time.Now()
	return &domain.Session{
		ID:        "test-session-id",
		Title:     "Test Investigation",
		StartTime: now,
		EndTime:   &now,
		Notes:     "Test notes",
		Status:    domain.SessionStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		Location: domain.Location{
			Latitude:    37.7749,
			Longitude:   -122.4194,
			Address:     "Test Location",
			Description: "Test Description",
			Venue:       "Test Venue",
		},
		Environmental: domain.Environmental{
			Temperature: 20.5,
			Humidity:    60.0,
			Pressure:    1013.25,
			EMFLevel:    0.5,
			LightLevel:  300.0,
			NoiseLevel:  45.0,
		},
	}
}

func createTestEVP() *domain.EVPRecording {
	now := time.Now()
	return &domain.EVPRecording{
		ID:             "test-evp-id",
		SessionID:      "test-session-id",
		FilePath:       "sessions/test-session-id/audio/evp.wav",
		Duration:       5.2,
		Timestamp:      now,
		WaveformData:   []float64{0.1, 0.2, 0.3, 0.2, 0.1},
		ProcessedPath:  "sessions/test-session-id/processed/evp_processed.wav",
		Annotations:    []string{"anomaly at 2.1s", "possible voice"},
		Quality:        domain.EVPQualityGood,
		DetectionLevel: 0.75,
		CreatedAt:      now,
	}
}

// Session Repository Tests

func TestSQLiteSessionRepository_Create_ValidSession_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()

	// Act
	err := repo.Create(context.Background(), session)

	// Assert
	assert.NoError(t, err)

	// Verify insertion
	retrieved, err := repo.GetByID(context.Background(), session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, retrieved.ID)
	assert.Equal(t, session.Title, retrieved.Title)
	assert.Equal(t, session.Status, retrieved.Status)
}

func TestSQLiteSessionRepository_GetByID_NonExistentID_Error(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)

	// Act
	_, err := repo.GetByID(context.Background(), "non-existent-id")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestSQLiteSessionRepository_GetAll_WithPagination_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)

	// Create test sessions
	sessions := []*domain.Session{
		createTestSession(),
		func() *domain.Session {
			s := createTestSession()
			s.ID = "session-2"
			s.Title = "Second Session"
			return s
		}(),
	}

	for _, session := range sessions {
		err := repo.Create(context.Background(), session)
		require.NoError(t, err)
	}

	// Act
	retrieved, err := repo.GetAll(context.Background(), 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)
}

func TestSQLiteSessionRepository_GetByStatus_ValidStatus_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()
	session.Status = domain.SessionStatusPaused

	err := repo.Create(context.Background(), session)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByStatus(context.Background(), domain.SessionStatusPaused)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, session.ID, retrieved[0].ID)
}

func TestSQLiteSessionRepository_GetByDateRange_ValidRange_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()

	err := repo.Create(context.Background(), session)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByDateRange(context.Background(),
		session.StartTime.Add(-time.Hour),
		session.StartTime.Add(time.Hour))

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, session.ID, retrieved[0].ID)
}

func TestSQLiteSessionRepository_Update_ValidSession_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()

	err := repo.Create(context.Background(), session)
	require.NoError(t, err)

	// Update session
	session.Title = "Updated Title"
	session.Notes = "Updated notes"

	// Act
	err = repo.Update(context.Background(), session)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), session.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, "Updated notes", retrieved.Notes)
}

func TestSQLiteSessionRepository_Delete_ValidID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()

	err := repo.Create(context.Background(), session)
	require.NoError(t, err)

	// Act
	err = repo.Delete(context.Background(), session.ID)

	// Assert
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), session.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// EVP Repository Tests

func TestSQLiteEVPRepository_Create_ValidEVP_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()

	// Act
	err := repo.Create(context.Background(), evp)

	// Assert
	assert.NoError(t, err)

	// Verify insertion
	retrieved, err := repo.GetByID(context.Background(), evp.ID)
	assert.NoError(t, err)
	assert.Equal(t, evp.ID, retrieved.ID)
	assert.Equal(t, evp.SessionID, retrieved.SessionID)
	assert.Equal(t, evp.Quality, retrieved.Quality)
}

func TestSQLiteEVPRepository_GetBySessionID_ValidSessionID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()

	err := repo.Create(context.Background(), evp)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetBySessionID(context.Background(), evp.SessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, evp.ID, retrieved[0].ID)
}

func TestSQLiteEVPRepository_GetByQuality_ValidQuality_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()
	evp.Quality = domain.EVPQualityExcellent

	err := repo.Create(context.Background(), evp)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByQuality(context.Background(), domain.EVPQualityExcellent)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, evp.ID, retrieved[0].ID)
}

func TestSQLiteEVPRepository_GetByDetectionLevel_ValidLevel_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()
	evp.DetectionLevel = 0.85

	err := repo.Create(context.Background(), evp)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByDetectionLevel(context.Background(), 0.8)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, evp.ID, retrieved[0].ID)
}

func TestSQLiteEVPRepository_Update_ValidEVP_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()

	err := repo.Create(context.Background(), evp)
	require.NoError(t, err)

	// Update EVP
	evp.Quality = domain.EVPQualityExcellent
	evp.DetectionLevel = 0.95

	// Act
	err = repo.Update(context.Background(), evp)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), evp.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.EVPQualityExcellent, retrieved.Quality)
	assert.Equal(t, 0.95, retrieved.DetectionLevel)
}

func TestSQLiteEVPRepository_Delete_ValidID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()

	err := repo.Create(context.Background(), evp)
	require.NoError(t, err)

	// Act
	err = repo.Delete(context.Background(), evp.ID)

	// Assert
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), evp.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// Repository Transaction Tests

func TestSQLiteSessionRepository_TransactionRollback_Error(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()

	// Begin transaction
	tx, err := db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	defer tx.Rollback()

	// Act - Try to insert with invalid data
	invalidSession := createTestSession()
	invalidSession.ID = "" // Invalid empty ID

	err = repo.Create(context.Background(), invalidSession)
	assert.Error(t, err)

	// Rollback transaction
	err = tx.Rollback()
	assert.NoError(t, err)

	// Assert - Session should not exist
	_, err = repo.GetByID(context.Background(), session.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// Concurrency Tests

func TestSQLiteSessionRepository_ConcurrentOperations_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)

	// Act - Create multiple sessions concurrently
	done := make(chan bool, 5)

	for index := 0; index < 5; index++ {
		go func(idx int) {
			session := createTestSession()
			session.ID = fmt.Sprintf("session-%d", idx)
			session.Title = fmt.Sprintf("Session %d", idx)

			err := repo.Create(context.Background(), session)
			assert.NoError(t, err)
			done <- true
		}(index)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Assert - All sessions should be created
	sessions, err := repo.GetAll(context.Background(), 100, 0)
	assert.NoError(t, err)
	assert.Len(t, sessions, 5)
}
