package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/myideascope/otherside/internal/domain"
)

// SQLiteSessionRepository implements SessionRepository using SQLite
type SQLiteSessionRepository struct {
	db *sql.DB
}

// NewSQLiteSessionRepository creates a new SQLite session repository
func NewSQLiteSessionRepository(db *sql.DB) *SQLiteSessionRepository {
	return &SQLiteSessionRepository{db: db}
}

// Create creates a new session
func (r *SQLiteSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (
			id, title, location_latitude, location_longitude, location_address, 
			location_description, location_venue, start_time, end_time, notes,
			env_temperature, env_humidity, env_pressure, env_emf_level, 
			env_light_level, env_noise_level, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.Title, session.Location.Latitude, session.Location.Longitude,
		session.Location.Address, session.Location.Description, session.Location.Venue,
		session.StartTime, session.EndTime, session.Notes,
		session.Environmental.Temperature, session.Environmental.Humidity,
		session.Environmental.Pressure, session.Environmental.EMFLevel,
		session.Environmental.LightLevel, session.Environmental.NoiseLevel,
		session.Status, session.CreatedAt, session.UpdatedAt,
	)

	return err
}

// GetByID retrieves a session by ID
func (r *SQLiteSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	query := `
		SELECT id, title, location_latitude, location_longitude, location_address,
			location_description, location_venue, start_time, end_time, notes,
			env_temperature, env_humidity, env_pressure, env_emf_level,
			env_light_level, env_noise_level, status, created_at, updated_at
		FROM sessions WHERE id = ?`

	var session domain.Session
	var endTime sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.Title, &session.Location.Latitude, &session.Location.Longitude,
		&session.Location.Address, &session.Location.Description, &session.Location.Venue,
		&session.StartTime, &endTime, &session.Notes,
		&session.Environmental.Temperature, &session.Environmental.Humidity,
		&session.Environmental.Pressure, &session.Environmental.EMFLevel,
		&session.Environmental.LightLevel, &session.Environmental.NoiseLevel,
		&session.Status, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if endTime.Valid {
		session.EndTime = &endTime.Time
	}

	return &session, nil
}

// GetAll retrieves all sessions with pagination
func (r *SQLiteSessionRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Session, error) {
	query := `
		SELECT id, title, location_latitude, location_longitude, location_address,
			location_description, location_venue, start_time, end_time, notes,
			env_temperature, env_humidity, env_pressure, env_emf_level,
			env_light_level, env_noise_level, status, created_at, updated_at
		FROM sessions ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var session domain.Session
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID, &session.Title, &session.Location.Latitude, &session.Location.Longitude,
			&session.Location.Address, &session.Location.Description, &session.Location.Venue,
			&session.StartTime, &endTime, &session.Notes,
			&session.Environmental.Temperature, &session.Environmental.Humidity,
			&session.Environmental.Pressure, &session.Environmental.EMFLevel,
			&session.Environmental.LightLevel, &session.Environmental.NoiseLevel,
			&session.Status, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

// GetByStatus retrieves sessions by status
func (r *SQLiteSessionRepository) GetByStatus(ctx context.Context, status domain.SessionStatus) ([]*domain.Session, error) {
	query := `
		SELECT id, title, location_latitude, location_longitude, location_address,
			location_description, location_venue, start_time, end_time, notes,
			env_temperature, env_humidity, env_pressure, env_emf_level,
			env_light_level, env_noise_level, status, created_at, updated_at
		FROM sessions WHERE status = ? ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var session domain.Session
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID, &session.Title, &session.Location.Latitude, &session.Location.Longitude,
			&session.Location.Address, &session.Location.Description, &session.Location.Venue,
			&session.StartTime, &endTime, &session.Notes,
			&session.Environmental.Temperature, &session.Environmental.Humidity,
			&session.Environmental.Pressure, &session.Environmental.EMFLevel,
			&session.Environmental.LightLevel, &session.Environmental.NoiseLevel,
			&session.Status, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

// Update updates a session
func (r *SQLiteSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions SET
			title = ?, location_latitude = ?, location_longitude = ?, 
			location_address = ?, location_description = ?, location_venue = ?,
			start_time = ?, end_time = ?, notes = ?,
			env_temperature = ?, env_humidity = ?, env_pressure = ?,
			env_emf_level = ?, env_light_level = ?, env_noise_level = ?,
			status = ?, updated_at = ?
		WHERE id = ?`

	session.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		session.Title, session.Location.Latitude, session.Location.Longitude,
		session.Location.Address, session.Location.Description, session.Location.Venue,
		session.StartTime, session.EndTime, session.Notes,
		session.Environmental.Temperature, session.Environmental.Humidity,
		session.Environmental.Pressure, session.Environmental.EMFLevel,
		session.Environmental.LightLevel, session.Environmental.NoiseLevel,
		session.Status, session.UpdatedAt, session.ID,
	)

	return err
}

// Delete deletes a session
func (r *SQLiteSessionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByDateRange retrieves sessions within a date range
func (r *SQLiteSessionRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Session, error) {
	query := `
		SELECT id, title, location_latitude, location_longitude, location_address,
			location_description, location_venue, start_time, end_time, notes,
			env_temperature, env_humidity, env_pressure, env_emf_level,
			env_light_level, env_noise_level, status, created_at, updated_at
		FROM sessions 
		WHERE start_time >= ? AND start_time <= ?
		ORDER BY start_time DESC`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		var session domain.Session
		var endTime sql.NullTime

		err := rows.Scan(
			&session.ID, &session.Title, &session.Location.Latitude, &session.Location.Longitude,
			&session.Location.Address, &session.Location.Description, &session.Location.Venue,
			&session.StartTime, &endTime, &session.Notes,
			&session.Environmental.Temperature, &session.Environmental.Humidity,
			&session.Environmental.Pressure, &session.Environmental.EMFLevel,
			&session.Environmental.LightLevel, &session.Environmental.NoiseLevel,
			&session.Status, &session.CreatedAt, &session.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if endTime.Valid {
			session.EndTime = &endTime.Time
		}

		sessions = append(sessions, &session)
	}

	return sessions, rows.Err()
}

// SQLiteEVPRepository implements EVPRepository using SQLite
type SQLiteEVPRepository struct {
	db *sql.DB
}

// NewSQLiteEVPRepository creates a new SQLite EVP repository
func NewSQLiteEVPRepository(db *sql.DB) *SQLiteEVPRepository {
	return &SQLiteEVPRepository{db: db}
}

// Create creates a new EVP recording
func (r *SQLiteEVPRepository) Create(ctx context.Context, evp *domain.EVPRecording) error {
	waveformJSON, _ := json.Marshal(evp.WaveformData)
	annotationsJSON, _ := json.Marshal(evp.Annotations)

	query := `
		INSERT INTO evp_recordings (
			id, session_id, file_path, duration, timestamp, waveform_data,
			processed_path, annotations, quality, detection_level, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		evp.ID, evp.SessionID, evp.FilePath, evp.Duration, evp.Timestamp,
		waveformJSON, evp.ProcessedPath, annotationsJSON,
		evp.Quality, evp.DetectionLevel, evp.CreatedAt,
	)

	return err
}

// GetByID retrieves an EVP recording by ID
func (r *SQLiteEVPRepository) GetByID(ctx context.Context, id string) (*domain.EVPRecording, error) {
	query := `
		SELECT id, session_id, file_path, duration, timestamp, waveform_data,
			processed_path, annotations, quality, detection_level, created_at
		FROM evp_recordings WHERE id = ?`

	var evp domain.EVPRecording
	var waveformJSON, annotationsJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&evp.ID, &evp.SessionID, &evp.FilePath, &evp.Duration, &evp.Timestamp,
		&waveformJSON, &evp.ProcessedPath, &annotationsJSON,
		&evp.Quality, &evp.DetectionLevel, &evp.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(waveformJSON), &evp.WaveformData)
	json.Unmarshal([]byte(annotationsJSON), &evp.Annotations)

	return &evp, nil
}

// GetBySessionID retrieves EVP recordings by session ID
func (r *SQLiteEVPRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.EVPRecording, error) {
	query := `
		SELECT id, session_id, file_path, duration, timestamp, waveform_data,
			processed_path, annotations, quality, detection_level, created_at
		FROM evp_recordings WHERE session_id = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evps []*domain.EVPRecording
	for rows.Next() {
		var evp domain.EVPRecording
		var waveformJSON, annotationsJSON string

		err := rows.Scan(
			&evp.ID, &evp.SessionID, &evp.FilePath, &evp.Duration, &evp.Timestamp,
			&waveformJSON, &evp.ProcessedPath, &annotationsJSON,
			&evp.Quality, &evp.DetectionLevel, &evp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(waveformJSON), &evp.WaveformData)
		json.Unmarshal([]byte(annotationsJSON), &evp.Annotations)

		evps = append(evps, &evp)
	}

	return evps, rows.Err()
}

// Update updates an EVP recording
func (r *SQLiteEVPRepository) Update(ctx context.Context, evp *domain.EVPRecording) error {
	waveformJSON, _ := json.Marshal(evp.WaveformData)
	annotationsJSON, _ := json.Marshal(evp.Annotations)

	query := `
		UPDATE evp_recordings SET
			file_path = ?, duration = ?, timestamp = ?, waveform_data = ?,
			processed_path = ?, annotations = ?, quality = ?, detection_level = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		evp.FilePath, evp.Duration, evp.Timestamp, waveformJSON,
		evp.ProcessedPath, annotationsJSON, evp.Quality, evp.DetectionLevel,
		evp.ID,
	)

	return err
}

// Delete deletes an EVP recording
func (r *SQLiteEVPRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM evp_recordings WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByQuality retrieves EVP recordings by quality
func (r *SQLiteEVPRepository) GetByQuality(ctx context.Context, quality domain.EVPQuality) ([]*domain.EVPRecording, error) {
	query := `
		SELECT id, session_id, file_path, duration, timestamp, waveform_data,
			processed_path, annotations, quality, detection_level, created_at
		FROM evp_recordings WHERE quality = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, quality)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evps []*domain.EVPRecording
	for rows.Next() {
		var evp domain.EVPRecording
		var waveformJSON, annotationsJSON string

		err := rows.Scan(
			&evp.ID, &evp.SessionID, &evp.FilePath, &evp.Duration, &evp.Timestamp,
			&waveformJSON, &evp.ProcessedPath, &annotationsJSON,
			&evp.Quality, &evp.DetectionLevel, &evp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(waveformJSON), &evp.WaveformData)
		json.Unmarshal([]byte(annotationsJSON), &evp.Annotations)

		evps = append(evps, &evp)
	}

	return evps, rows.Err()
}

// GetByDetectionLevel retrieves EVP recordings by detection level
func (r *SQLiteEVPRepository) GetByDetectionLevel(ctx context.Context, minLevel float64) ([]*domain.EVPRecording, error) {
	query := `
		SELECT id, session_id, file_path, duration, timestamp, waveform_data,
			processed_path, annotations, quality, detection_level, created_at
		FROM evp_recordings WHERE detection_level >= ? ORDER BY detection_level DESC`

	rows, err := r.db.QueryContext(ctx, query, minLevel)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evps []*domain.EVPRecording
	for rows.Next() {
		var evp domain.EVPRecording
		var waveformJSON, annotationsJSON string

		err := rows.Scan(
			&evp.ID, &evp.SessionID, &evp.FilePath, &evp.Duration, &evp.Timestamp,
			&waveformJSON, &evp.ProcessedPath, &annotationsJSON,
			&evp.Quality, &evp.DetectionLevel, &evp.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(waveformJSON), &evp.WaveformData)
		json.Unmarshal([]byte(annotationsJSON), &evp.Annotations)

		evps = append(evps, &evp)
	}

	return evps, rows.Err()
}

// Database initialization and migration functions

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Enable foreign key support
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	return db, nil
}

// RunMigrations runs database schema migrations
func RunMigrations(db *sql.DB) error {
	// Read schema file
	schemaBytes, err := os.ReadFile("configs/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}
