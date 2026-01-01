package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/myideascope/otherside/internal/domain"
)

// SQLiteVOXRepository implements VOXRepository using SQLite
type SQLiteVOXRepository struct {
	db *sql.DB
}

// NewSQLiteVOXRepository creates a new SQLite VOX repository
func NewSQLiteVOXRepository(db *sql.DB) *SQLiteVOXRepository {
	return &SQLiteVOXRepository{db: db}
}

// Create creates a new VOX event
func (r *SQLiteVOXRepository) Create(ctx context.Context, vox *domain.VOXEvent) error {
	frequencyJSON, _ := json.Marshal(vox.FrequencyData)

	query := `
		INSERT INTO vox_events (
			id, session_id, timestamp, generated_text, phonetic_bank, frequency_data,
			trigger_strength, language_pack, modulation_type, user_response, response_delay, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		vox.ID, vox.SessionID, vox.Timestamp, vox.GeneratedText, vox.PhoneticBank,
		frequencyJSON, vox.TriggerStrength, vox.LanguagePack, vox.ModulationType,
		vox.UserResponse, vox.ResponseDelay, vox.CreatedAt,
	)

	return err
}

// GetByID retrieves a VOX event by ID
func (r *SQLiteVOXRepository) GetByID(ctx context.Context, id string) (*domain.VOXEvent, error) {
	query := `
		SELECT id, session_id, timestamp, generated_text, phonetic_bank, frequency_data,
			trigger_strength, language_pack, modulation_type, user_response, response_delay, created_at
		FROM vox_events WHERE id = ?`

	var vox domain.VOXEvent
	var frequencyJSON string
	var userResponse, responseDelay sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&vox.ID, &vox.SessionID, &vox.Timestamp, &vox.GeneratedText, &vox.PhoneticBank,
		&frequencyJSON, &vox.TriggerStrength, &vox.LanguagePack, &vox.ModulationType,
		&userResponse, &responseDelay, &vox.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(frequencyJSON), &vox.FrequencyData)

	if userResponse.Valid {
		vox.UserResponse = userResponse.String
	}
	if responseDelay.Valid {
		if delay, err := parseFloat64(responseDelay.String); err == nil {
			vox.ResponseDelay = delay
		}
	}

	return &vox, nil
}

// GetBySessionID retrieves VOX events by session ID
func (r *SQLiteVOXRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.VOXEvent, error) {
	query := `
		SELECT id, session_id, timestamp, generated_text, phonetic_bank, frequency_data,
			trigger_strength, language_pack, modulation_type, user_response, response_delay, created_at
		FROM vox_events WHERE session_id = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voxEvents []*domain.VOXEvent
	for rows.Next() {
		var vox domain.VOXEvent
		var frequencyJSON string
		var userResponse, responseDelay sql.NullString

		err := rows.Scan(
			&vox.ID, &vox.SessionID, &vox.Timestamp, &vox.GeneratedText, &vox.PhoneticBank,
			&frequencyJSON, &vox.TriggerStrength, &vox.LanguagePack, &vox.ModulationType,
			&userResponse, &responseDelay, &vox.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(frequencyJSON), &vox.FrequencyData)

		if userResponse.Valid {
			vox.UserResponse = userResponse.String
		}
		if responseDelay.Valid {
			if delay, err := parseFloat64(responseDelay.String); err == nil {
				vox.ResponseDelay = delay
			}
		}

		voxEvents = append(voxEvents, &vox)
	}

	return voxEvents, rows.Err()
}

// Update updates a VOX event
func (r *SQLiteVOXRepository) Update(ctx context.Context, vox *domain.VOXEvent) error {
	frequencyJSON, _ := json.Marshal(vox.FrequencyData)

	query := `
		UPDATE vox_events SET
			timestamp = ?, generated_text = ?, phonetic_bank = ?, frequency_data = ?,
			trigger_strength = ?, language_pack = ?, modulation_type = ?,
			user_response = ?, response_delay = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		vox.Timestamp, vox.GeneratedText, vox.PhoneticBank, frequencyJSON,
		vox.TriggerStrength, vox.LanguagePack, vox.ModulationType,
		vox.UserResponse, vox.ResponseDelay, vox.ID,
	)

	return err
}

// Delete deletes a VOX event
func (r *SQLiteVOXRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM vox_events WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByLanguagePack retrieves VOX events by language pack
func (r *SQLiteVOXRepository) GetByLanguagePack(ctx context.Context, languagePack string) ([]*domain.VOXEvent, error) {
	query := `
		SELECT id, session_id, timestamp, generated_text, phonetic_bank, frequency_data,
			trigger_strength, language_pack, modulation_type, user_response, response_delay, created_at
		FROM vox_events WHERE language_pack = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, languagePack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voxEvents []*domain.VOXEvent
	for rows.Next() {
		var vox domain.VOXEvent
		var frequencyJSON string
		var userResponse, responseDelay sql.NullString

		err := rows.Scan(
			&vox.ID, &vox.SessionID, &vox.Timestamp, &vox.GeneratedText, &vox.PhoneticBank,
			&frequencyJSON, &vox.TriggerStrength, &vox.LanguagePack, &vox.ModulationType,
			&userResponse, &responseDelay, &vox.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(frequencyJSON), &vox.FrequencyData)

		if userResponse.Valid {
			vox.UserResponse = userResponse.String
		}
		if responseDelay.Valid {
			if delay, err := parseFloat64(responseDelay.String); err == nil {
				vox.ResponseDelay = delay
			}
		}

		voxEvents = append(voxEvents, &vox)
	}

	return voxEvents, rows.Err()
}

// GetByTriggerStrength retrieves VOX events by minimum trigger strength
func (r *SQLiteVOXRepository) GetByTriggerStrength(ctx context.Context, minStrength float64) ([]*domain.VOXEvent, error) {
	query := `
		SELECT id, session_id, timestamp, generated_text, phonetic_bank, frequency_data,
			trigger_strength, language_pack, modulation_type, user_response, response_delay, created_at
		FROM vox_events WHERE trigger_strength >= ? ORDER BY trigger_strength DESC`

	rows, err := r.db.QueryContext(ctx, query, minStrength)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voxEvents []*domain.VOXEvent
	for rows.Next() {
		var vox domain.VOXEvent
		var frequencyJSON string
		var userResponse, responseDelay sql.NullString

		err := rows.Scan(
			&vox.ID, &vox.SessionID, &vox.Timestamp, &vox.GeneratedText, &vox.PhoneticBank,
			&frequencyJSON, &vox.TriggerStrength, &vox.LanguagePack, &vox.ModulationType,
			&userResponse, &responseDelay, &vox.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(frequencyJSON), &vox.FrequencyData)

		if userResponse.Valid {
			vox.UserResponse = userResponse.String
		}
		if responseDelay.Valid {
			if delay, err := parseFloat64(responseDelay.String); err == nil {
				vox.ResponseDelay = delay
			}
		}

		voxEvents = append(voxEvents, &vox)
	}

	return voxEvents, rows.Err()
}

// SQLiteRadarRepository implements RadarRepository using SQLite
type SQLiteRadarRepository struct {
	db *sql.DB
}

// NewSQLiteRadarRepository creates a new SQLite radar repository
func NewSQLiteRadarRepository(db *sql.DB) *SQLiteRadarRepository {
	return &SQLiteRadarRepository{db: db}
}

// Create creates a new radar event
func (r *SQLiteRadarRepository) Create(ctx context.Context, radar *domain.RadarEvent) error {
	movementJSON, _ := json.Marshal(radar.MovementTrail)

	query := `
		INSERT INTO radar_events (
			id, session_id, timestamp, position_x, position_y, position_z,
			strength, source_type, emf_reading, audio_anomaly, duration, movement_trail, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		radar.ID, radar.SessionID, radar.Timestamp, radar.Position.X, radar.Position.Y, radar.Position.Z,
		radar.Strength, radar.SourceType, radar.EMFReading, radar.AudioAnomaly, radar.Duration,
		movementJSON, radar.CreatedAt,
	)

	return err
}

// GetByID retrieves a radar event by ID
func (r *SQLiteRadarRepository) GetByID(ctx context.Context, id string) (*domain.RadarEvent, error) {
	query := `
		SELECT id, session_id, timestamp, position_x, position_y, position_z,
			strength, source_type, emf_reading, audio_anomaly, duration, movement_trail, created_at
		FROM radar_events WHERE id = ?`

	var radar domain.RadarEvent
	var movementJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&radar.ID, &radar.SessionID, &radar.Timestamp, &radar.Position.X, &radar.Position.Y, &radar.Position.Z,
		&radar.Strength, &radar.SourceType, &radar.EMFReading, &radar.AudioAnomaly, &radar.Duration,
		&movementJSON, &radar.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(movementJSON), &radar.MovementTrail)

	return &radar, nil
}

// GetBySessionID retrieves radar events by session ID
func (r *SQLiteRadarRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.RadarEvent, error) {
	query := `
		SELECT id, session_id, timestamp, position_x, position_y, position_z,
			strength, source_type, emf_reading, audio_anomaly, duration, movement_trail, created_at
		FROM radar_events WHERE session_id = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var radarEvents []*domain.RadarEvent
	for rows.Next() {
		var radar domain.RadarEvent
		var movementJSON string

		err := rows.Scan(
			&radar.ID, &radar.SessionID, &radar.Timestamp, &radar.Position.X, &radar.Position.Y, &radar.Position.Z,
			&radar.Strength, &radar.SourceType, &radar.EMFReading, &radar.AudioAnomaly, &radar.Duration,
			&movementJSON, &radar.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(movementJSON), &radar.MovementTrail)

		radarEvents = append(radarEvents, &radar)
	}

	return radarEvents, rows.Err()
}

// Update updates a radar event
func (r *SQLiteRadarRepository) Update(ctx context.Context, radar *domain.RadarEvent) error {
	movementJSON, _ := json.Marshal(radar.MovementTrail)

	query := `
		UPDATE radar_events SET
			timestamp = ?, position_x = ?, position_y = ?, position_z = ?,
			strength = ?, source_type = ?, emf_reading = ?, audio_anomaly = ?,
			duration = ?, movement_trail = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		radar.Timestamp, radar.Position.X, radar.Position.Y, radar.Position.Z,
		radar.Strength, radar.SourceType, radar.EMFReading, radar.AudioAnomaly,
		radar.Duration, movementJSON, radar.ID,
	)

	return err
}

// Delete deletes a radar event
func (r *SQLiteRadarRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM radar_events WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetBySourceType retrieves radar events by source type
func (r *SQLiteRadarRepository) GetBySourceType(ctx context.Context, sourceType domain.SourceType) ([]*domain.RadarEvent, error) {
	query := `
		SELECT id, session_id, timestamp, position_x, position_y, position_z,
			strength, source_type, emf_reading, audio_anomaly, duration, movement_trail, created_at
		FROM radar_events WHERE source_type = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, sourceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var radarEvents []*domain.RadarEvent
	for rows.Next() {
		var radar domain.RadarEvent
		var movementJSON string

		err := rows.Scan(
			&radar.ID, &radar.SessionID, &radar.Timestamp, &radar.Position.X, &radar.Position.Y, &radar.Position.Z,
			&radar.Strength, &radar.SourceType, &radar.EMFReading, &radar.AudioAnomaly, &radar.Duration,
			&movementJSON, &radar.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(movementJSON), &radar.MovementTrail)

		radarEvents = append(radarEvents, &radar)
	}

	return radarEvents, rows.Err()
}

// GetByStrengthRange retrieves radar events by strength range
func (r *SQLiteRadarRepository) GetByStrengthRange(ctx context.Context, minStrength, maxStrength float64) ([]*domain.RadarEvent, error) {
	query := `
		SELECT id, session_id, timestamp, position_x, position_y, position_z,
			strength, source_type, emf_reading, audio_anomaly, duration, movement_trail, created_at
		FROM radar_events WHERE strength >= ? AND strength <= ? ORDER BY strength DESC`

	rows, err := r.db.QueryContext(ctx, query, minStrength, maxStrength)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var radarEvents []*domain.RadarEvent
	for rows.Next() {
		var radar domain.RadarEvent
		var movementJSON string

		err := rows.Scan(
			&radar.ID, &radar.SessionID, &radar.Timestamp, &radar.Position.X, &radar.Position.Y, &radar.Position.Z,
			&radar.Strength, &radar.SourceType, &radar.EMFReading, &radar.AudioAnomaly, &radar.Duration,
			&movementJSON, &radar.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(movementJSON), &radar.MovementTrail)

		radarEvents = append(radarEvents, &radar)
	}

	return radarEvents, rows.Err()
}

// Helper function to parse float64 from string
func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
