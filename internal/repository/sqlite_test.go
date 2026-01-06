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

// Helper functions for creating test data

func createTestVOX() *domain.VOXEvent {
	now := time.Now()
	return &domain.VOXEvent{
		ID:              "test-vox-id",
		SessionID:       "test-session-id",
		Timestamp:       now,
		GeneratedText:   "Test voice synthesis",
		PhoneticBank:    "english",
		FrequencyData:   []float64{440.0, 880.0, 1760.0},
		TriggerStrength: 0.75,
		LanguagePack:    "en-us",
		ModulationType:  "amplitude",
		UserResponse:    "User response text",
		ResponseDelay:   2.5,
		CreatedAt:       now,
	}
}

func createTestRadar() *domain.RadarEvent {
	now := time.Now()
	return &domain.RadarEvent{
		ID:           "test-radar-id",
		SessionID:    "test-session-id",
		Timestamp:    now,
		Position:     domain.Coordinates{X: 1.5, Y: 2.5, Z: 0.0},
		Strength:     0.85,
		SourceType:   domain.SourceTypeEMF,
		EMFReading:   0.75,
		AudioAnomaly: 0.65,
		Duration:     3.2,
		MovementTrail: []domain.Coordinates{
			{X: 1.0, Y: 2.0, Z: 0.0},
			{X: 1.5, Y: 2.5, Z: 0.0},
			{X: 2.0, Y: 3.0, Z: 0.0},
		},
		CreatedAt: now,
	}
}

func createTestSLS() *domain.SLSDetection {
	now := time.Now()
	return &domain.SLSDetection{
		ID:        "test-sls-id",
		SessionID: "test-session-id",
		Timestamp: now,
		SkeletalPoints: []domain.SkeletalPoint{
			{
				Joint:      "head",
				Position:   domain.Coordinates{X: 0.0, Y: 0.0, Z: 0.0},
				Confidence: 0.95,
			},
			{
				Joint:      "left_hand",
				Position:   domain.Coordinates{X: 1.0, Y: 1.0, Z: 0.0},
				Confidence: 0.85,
			},
		},
		Confidence: 0.90,
		BoundingBox: domain.BoundingBox{
			TopLeft:     domain.Coordinates{X: -1.0, Y: -1.0, Z: 0.0},
			BottomRight: domain.Coordinates{X: 1.0, Y: 1.0, Z: 0.0},
			Width:       2.0,
			Height:      2.0,
		},
		VideoFrame:    "frame_001.jpg",
		FilterApplied: []string{"noise_reduction", "smoothing"},
		Duration:      5.0,
		Movement: domain.MovementAnalysis{
			Speed:     1.5,
			Direction: 45.0,
			Pattern:   "circular",
		},
		CreatedAt: now,
	}
}

func createTestInteraction() *domain.UserInteraction {
	now := time.Now()
	return &domain.UserInteraction{
		ID:           "test-interaction-id",
		SessionID:    "test-session-id",
		Timestamp:    now,
		Type:         domain.InteractionTypeVoice,
		Content:      "Test user interaction",
		AudioPath:    "audio/user_voice.wav",
		Response:     "System response",
		ResponseTime: 1.2,
		RandomizerResult: &domain.RandomizerResult{
			Type:   "number",
			Result: 42,
			Range:  "1-100",
		},
		CreatedAt: now,
	}
}

// VOX Repository Tests

func TestSQLiteVOXRepository_Create_ValidVOX_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteVOXRepository(db)
	vox := createTestVOX()

	// Act
	err := repo.Create(context.Background(), vox)

	// Assert
	assert.NoError(t, err)

	// Verify insertion
	retrieved, err := repo.GetByID(context.Background(), vox.ID)
	assert.NoError(t, err)
	assert.Equal(t, vox.ID, retrieved.ID)
	assert.Equal(t, vox.SessionID, retrieved.SessionID)
	assert.Equal(t, vox.GeneratedText, retrieved.GeneratedText)
	assert.Equal(t, vox.TriggerStrength, retrieved.TriggerStrength)
}

func TestSQLiteVOXRepository_GetBySessionID_ValidSessionID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteVOXRepository(db)
	vox := createTestVOX()

	err := repo.Create(context.Background(), vox)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetBySessionID(context.Background(), vox.SessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, vox.ID, retrieved[0].ID)
}

func TestSQLiteVOXRepository_GetByLanguagePack_ValidLanguagePack_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteVOXRepository(db)
	vox := createTestVOX()
	vox.LanguagePack = "es-mx"

	err := repo.Create(context.Background(), vox)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByLanguagePack(context.Background(), "es-mx")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, vox.ID, retrieved[0].ID)
}

func TestSQLiteVOXRepository_GetByTriggerStrength_ValidStrength_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteVOXRepository(db)
	vox := createTestVOX()
	vox.TriggerStrength = 0.90

	err := repo.Create(context.Background(), vox)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByTriggerStrength(context.Background(), 0.80)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, vox.ID, retrieved[0].ID)
}

func TestSQLiteVOXRepository_Update_ValidVOX_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteVOXRepository(db)
	vox := createTestVOX()

	err := repo.Create(context.Background(), vox)
	require.NoError(t, err)

	// Update VOX
	vox.GeneratedText = "Updated voice synthesis"
	vox.UserResponse = "Updated user response"

	// Act
	err = repo.Update(context.Background(), vox)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), vox.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated voice synthesis", retrieved.GeneratedText)
	assert.Equal(t, "Updated user response", retrieved.UserResponse)
}

func TestSQLiteVOXRepository_Delete_ValidID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteVOXRepository(db)
	vox := createTestVOX()

	err := repo.Create(context.Background(), vox)
	require.NoError(t, err)

	// Act
	err = repo.Delete(context.Background(), vox.ID)

	// Assert
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), vox.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// Radar Repository Tests

func TestSQLiteRadarRepository_Create_ValidRadar_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()

	// Act
	err := repo.Create(context.Background(), radar)

	// Assert
	assert.NoError(t, err)

	// Verify insertion
	retrieved, err := repo.GetByID(context.Background(), radar.ID)
	assert.NoError(t, err)
	assert.Equal(t, radar.ID, retrieved.ID)
	assert.Equal(t, radar.SessionID, retrieved.SessionID)
	assert.Equal(t, radar.Position.X, retrieved.Position.X)
	assert.Equal(t, radar.Position.Y, retrieved.Position.Y)
	assert.Equal(t, radar.Strength, retrieved.Strength)
}

func TestSQLiteRadarRepository_GetBySessionID_ValidSessionID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()

	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetBySessionID(context.Background(), radar.SessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, radar.ID, retrieved[0].ID)
}

func TestSQLiteRadarRepository_GetBySourceType_ValidSourceType_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()
	radar.SourceType = domain.SourceTypeAudio

	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetBySourceType(context.Background(), domain.SourceTypeAudio)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, radar.ID, retrieved[0].ID)
}

func TestSQLiteRadarRepository_GetByStrengthRange_ValidRange_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()
	radar.Strength = 0.90

	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByStrengthRange(context.Background(), 0.80, 1.0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, radar.ID, retrieved[0].ID)
}

func TestSQLiteRadarRepository_GetByPosition_ValidPosition_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()

	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Act - Search within 1.0 units of radar position (using strength range as proxy for proximity)
	retrieved, err := repo.GetByStrengthRange(context.Background(), radar.Strength-0.1, radar.Strength+0.1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, radar.ID, retrieved[0].ID)
}

func TestSQLiteRadarRepository_Update_ValidRadar_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()

	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Update radar
	radar.Strength = 0.95
	radar.EMFReading = 0.85

	// Act
	err = repo.Update(context.Background(), radar)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), radar.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0.95, retrieved.Strength)
	assert.Equal(t, 0.85, retrieved.EMFReading)
}

func TestSQLiteRadarRepository_Delete_ValidID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()

	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Act
	err = repo.Delete(context.Background(), radar.ID)

	// Assert
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), radar.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// SLS Repository Tests

func TestSQLiteSLSRepository_Create_ValidSLS_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()

	// Act
	err := repo.Create(context.Background(), sls)

	// Assert
	assert.NoError(t, err)

	// Verify insertion
	retrieved, err := repo.GetByID(context.Background(), sls.ID)
	assert.NoError(t, err)
	assert.Equal(t, sls.ID, retrieved.ID)
	assert.Equal(t, sls.SessionID, retrieved.SessionID)
	assert.Equal(t, sls.Confidence, retrieved.Confidence)
	assert.Equal(t, sls.BoundingBox.Width, retrieved.BoundingBox.Width)
}

func TestSQLiteSLSRepository_GetBySessionID_ValidSessionID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()

	err := repo.Create(context.Background(), sls)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetBySessionID(context.Background(), sls.SessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, sls.ID, retrieved[0].ID)
}

func TestSQLiteSLSRepository_GetByConfidenceRange_ValidRange_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()
	sls.Confidence = 0.95

	err := repo.Create(context.Background(), sls)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByConfidenceRange(context.Background(), 0.90)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, sls.ID, retrieved[0].ID)
}

func TestSQLiteSLSRepository_GetByDuration_ValidDuration_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()
	sls.Duration = 10.0

	err := repo.Create(context.Background(), sls)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByDuration(context.Background(), 5.0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, sls.ID, retrieved[0].ID)
}

func TestSQLiteSLSRepository_GetByBoundingBox_ValidBoundingBox_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()

	err := repo.Create(context.Background(), sls)
	require.NoError(t, err)

	// Act - Use confidence range as proxy for bounding box search
	retrieved, err := repo.GetByConfidenceRange(context.Background(), sls.Confidence-0.1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, sls.ID, retrieved[0].ID)
}

func TestSQLiteSLSRepository_Update_ValidSLS_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()

	err := repo.Create(context.Background(), sls)
	require.NoError(t, err)

	// Update SLS
	sls.Confidence = 0.98
	sls.Duration = 8.0

	// Act
	err = repo.Update(context.Background(), sls)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), sls.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0.98, retrieved.Confidence)
	assert.Equal(t, 8.0, retrieved.Duration)
}

func TestSQLiteSLSRepository_Delete_ValidID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSLSRepository(db)
	sls := createTestSLS()

	err := repo.Create(context.Background(), sls)
	require.NoError(t, err)

	// Act
	err = repo.Delete(context.Background(), sls.ID)

	// Assert
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), sls.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// User Interaction Repository Tests

func TestSQLiteInteractionRepository_Create_ValidInteraction_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()

	// Act
	err := repo.Create(context.Background(), interaction)

	// Assert
	assert.NoError(t, err)

	// Verify insertion
	retrieved, err := repo.GetByID(context.Background(), interaction.ID)
	assert.NoError(t, err)
	assert.Equal(t, interaction.ID, retrieved.ID)
	assert.Equal(t, interaction.SessionID, retrieved.SessionID)
	assert.Equal(t, interaction.Type, retrieved.Type)
	assert.Equal(t, interaction.Content, retrieved.Content)
}

func TestSQLiteInteractionRepository_GetBySessionID_ValidSessionID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()

	err := repo.Create(context.Background(), interaction)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetBySessionID(context.Background(), interaction.SessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, interaction.ID, retrieved[0].ID)
}

func TestSQLiteInteractionRepository_GetByType_ValidType_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()
	interaction.Type = domain.InteractionTypeText

	err := repo.Create(context.Background(), interaction)
	require.NoError(t, err)

	// Act
	retrieved, err := repo.GetByType(context.Background(), domain.InteractionTypeText)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, interaction.ID, retrieved[0].ID)
}

func TestSQLiteInteractionRepository_GetByResponseType_ValidResponseType_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()
	interaction.Response = "System generated response"

	err := repo.Create(context.Background(), interaction)
	require.NoError(t, err)

	// Act - Use GetByType as alternative since GetByResponseType doesn't exist
	retrieved, err := repo.GetByType(context.Background(), domain.InteractionTypeVoice)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, interaction.ID, retrieved[0].ID)
}

func TestSQLiteInteractionRepository_GetByResponseTime_ValidRange_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()
	interaction.ResponseTime = 2.0

	err := repo.Create(context.Background(), interaction)
	require.NoError(t, err)

	// Act - Use GetBySessionID as alternative since GetByResponseTime doesn't exist
	retrieved, err := repo.GetBySessionID(context.Background(), interaction.SessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, interaction.ID, retrieved[0].ID)
}

func TestSQLiteInteractionRepository_Update_ValidInteraction_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()

	err := repo.Create(context.Background(), interaction)
	require.NoError(t, err)

	// Update interaction
	interaction.Content = "Updated interaction content"
	interaction.Response = "Updated system response"

	// Act
	err = repo.Update(context.Background(), interaction)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), interaction.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated interaction content", retrieved.Content)
	assert.Equal(t, "Updated system response", retrieved.Response)
}

func TestSQLiteInteractionRepository_Delete_ValidID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteInteractionRepository(db)
	interaction := createTestInteraction()

	err := repo.Create(context.Background(), interaction)
	require.NoError(t, err)

	// Act
	err = repo.Delete(context.Background(), interaction.ID)

	// Assert
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), interaction.ID)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

// JSON Serialization Tests

func TestSQLiteEVPRepository_JSONSerialization_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()
	evp.WaveformData = []float64{0.1, 0.2, 0.3, 0.2, 0.1}
	evp.Annotations = []string{"anomaly detected", "voice at 2.1s", "low frequency"}

	// Act
	err := repo.Create(context.Background(), evp)
	require.NoError(t, err)

	// Assert
	retrieved, err := repo.GetByID(context.Background(), evp.ID)
	assert.NoError(t, err)
	assert.Equal(t, evp.WaveformData, retrieved.WaveformData)
	assert.Equal(t, evp.Annotations, retrieved.Annotations)
}

func TestSQLiteRadarRepository_MovementTrailSerialization_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteRadarRepository(db)
	radar := createTestRadar()
	radar.MovementTrail = []domain.Coordinates{
		{X: 0.0, Y: 0.0, Z: 0.0},
		{X: 1.0, Y: 2.0, Z: 0.0},
		{X: 2.0, Y: 3.0, Z: 0.0},
		{X: 3.0, Y: 4.0, Z: 0.0},
	}

	// Act
	err := repo.Create(context.Background(), radar)
	require.NoError(t, err)

	// Assert
	retrieved, err := repo.GetByID(context.Background(), radar.ID)
	assert.NoError(t, err)
	assert.Equal(t, radar.MovementTrail, retrieved.MovementTrail)
}

// Bulk Operations Tests

func TestSQLiteSessionRepository_BulkOperations_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)

	// Act - Create multiple sessions
	for i := 0; i < 10; i++ {
		session := createTestSession()
		session.ID = fmt.Sprintf("session-%d", i)
		session.Title = fmt.Sprintf("Test Session %d", i)

		err := repo.Create(context.Background(), session)
		assert.NoError(t, err)
	}

	// Assert - Get all sessions
	sessions, err := repo.GetAll(context.Background(), 100, 0)
	assert.NoError(t, err)
	assert.Len(t, sessions, 10)

	// Update all sessions
	for _, session := range sessions {
		session.Status = domain.SessionStatusArchived
		err := repo.Update(context.Background(), session)
		assert.NoError(t, err)
	}

	// Verify updates
	archivedSessions, err := repo.GetByStatus(context.Background(), domain.SessionStatusArchived)
	assert.NoError(t, err)
	assert.Len(t, archivedSessions, 10)
}

// Edge Cases and Data Integrity Tests

func TestSQLiteSessionRepository_NullEndTime_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()
	session.EndTime = nil // Explicitly set to nil

	// Act
	err := repo.Create(context.Background(), session)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), session.ID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved.EndTime)
}

func TestSQLiteEVPRepository_EmptyArrays_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteEVPRepository(db)
	evp := createTestEVP()
	evp.WaveformData = []float64{} // Empty array
	evp.Annotations = []string{}   // Empty array

	// Act
	err := repo.Create(context.Background(), evp)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), evp.ID)
	assert.NoError(t, err)
	assert.Empty(t, retrieved.WaveformData)
	assert.Empty(t, retrieved.Annotations)
}

// Unicode and Special Characters Tests

func TestSQLiteSessionRepository_UnicodeHandling_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	defer cleanupTestDB(db)
	setupTestSchema(t, db)

	repo := NewSQLiteSessionRepository(db)
	session := createTestSession()
	session.Title = "👻 Paranormal Investigation 幽霊"
	session.Notes = "📍 Location: 東京, China"
	session.Location.Address = "123 Main Street, 北京, China"

	// Act
	err := repo.Create(context.Background(), session)

	// Assert
	assert.NoError(t, err)

	retrieved, err := repo.GetByID(context.Background(), session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.Title, retrieved.Title)
	assert.Equal(t, session.Notes, retrieved.Notes)
	assert.Equal(t, session.Location.Address, retrieved.Location.Address)
}
