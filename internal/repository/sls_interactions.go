package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/myideascope/otherside/internal/domain"
)

// SQLiteSLSRepository implements SLSRepository using SQLite
type SQLiteSLSRepository struct {
	db *sql.DB
}

// NewSQLiteSLSRepository creates a new SQLite SLS repository
func NewSQLiteSLSRepository(db *sql.DB) *SQLiteSLSRepository {
	return &SQLiteSLSRepository{db: db}
}

// Create creates a new SLS detection
func (r *SQLiteSLSRepository) Create(ctx context.Context, sls *domain.SLSDetection) error {
	skeletalJSON, _ := json.Marshal(sls.SkeletalPoints)
	filtersJSON, _ := json.Marshal(sls.FilterApplied)

	query := `
		INSERT INTO sls_detections (
			id, session_id, timestamp, skeletal_points, confidence,
			bounding_box_top_left_x, bounding_box_top_left_y,
			bounding_box_bottom_right_x, bounding_box_bottom_right_y,
			bounding_box_width, bounding_box_height, video_frame, filter_applied,
			duration, movement_speed, movement_direction, movement_pattern, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		sls.ID, sls.SessionID, sls.Timestamp, skeletalJSON, sls.Confidence,
		sls.BoundingBox.TopLeft.X, sls.BoundingBox.TopLeft.Y,
		sls.BoundingBox.BottomRight.X, sls.BoundingBox.BottomRight.Y,
		sls.BoundingBox.Width, sls.BoundingBox.Height, sls.VideoFrame, filtersJSON,
		sls.Duration, sls.Movement.Speed, sls.Movement.Direction, sls.Movement.Pattern,
		sls.CreatedAt,
	)

	return err
}

// GetByID retrieves an SLS detection by ID
func (r *SQLiteSLSRepository) GetByID(ctx context.Context, id string) (*domain.SLSDetection, error) {
	query := `
		SELECT id, session_id, timestamp, skeletal_points, confidence,
			bounding_box_top_left_x, bounding_box_top_left_y,
			bounding_box_bottom_right_x, bounding_box_bottom_right_y,
			bounding_box_width, bounding_box_height, video_frame, filter_applied,
			duration, movement_speed, movement_direction, movement_pattern, created_at
		FROM sls_detections WHERE id = ?`

	var sls domain.SLSDetection
	var skeletalJSON, filtersJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sls.ID, &sls.SessionID, &sls.Timestamp, &skeletalJSON, &sls.Confidence,
		&sls.BoundingBox.TopLeft.X, &sls.BoundingBox.TopLeft.Y,
		&sls.BoundingBox.BottomRight.X, &sls.BoundingBox.BottomRight.Y,
		&sls.BoundingBox.Width, &sls.BoundingBox.Height, &sls.VideoFrame, &filtersJSON,
		&sls.Duration, &sls.Movement.Speed, &sls.Movement.Direction, &sls.Movement.Pattern,
		&sls.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(skeletalJSON), &sls.SkeletalPoints)
	json.Unmarshal([]byte(filtersJSON), &sls.FilterApplied)

	return &sls, nil
}

// GetBySessionID retrieves SLS detections by session ID
func (r *SQLiteSLSRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.SLSDetection, error) {
	query := `
		SELECT id, session_id, timestamp, skeletal_points, confidence,
			bounding_box_top_left_x, bounding_box_top_left_y,
			bounding_box_bottom_right_x, bounding_box_bottom_right_y,
			bounding_box_width, bounding_box_height, video_frame, filter_applied,
			duration, movement_speed, movement_direction, movement_pattern, created_at
		FROM sls_detections WHERE session_id = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slsDetections []*domain.SLSDetection
	for rows.Next() {
		var sls domain.SLSDetection
		var skeletalJSON, filtersJSON string

		err := rows.Scan(
			&sls.ID, &sls.SessionID, &sls.Timestamp, &skeletalJSON, &sls.Confidence,
			&sls.BoundingBox.TopLeft.X, &sls.BoundingBox.TopLeft.Y,
			&sls.BoundingBox.BottomRight.X, &sls.BoundingBox.BottomRight.Y,
			&sls.BoundingBox.Width, &sls.BoundingBox.Height, &sls.VideoFrame, &filtersJSON,
			&sls.Duration, &sls.Movement.Speed, &sls.Movement.Direction, &sls.Movement.Pattern,
			&sls.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(skeletalJSON), &sls.SkeletalPoints)
		json.Unmarshal([]byte(filtersJSON), &sls.FilterApplied)

		slsDetections = append(slsDetections, &sls)
	}

	return slsDetections, rows.Err()
}

// Update updates an SLS detection
func (r *SQLiteSLSRepository) Update(ctx context.Context, sls *domain.SLSDetection) error {
	skeletalJSON, _ := json.Marshal(sls.SkeletalPoints)
	filtersJSON, _ := json.Marshal(sls.FilterApplied)

	query := `
		UPDATE sls_detections SET
			timestamp = ?, skeletal_points = ?, confidence = ?,
			bounding_box_top_left_x = ?, bounding_box_top_left_y = ?,
			bounding_box_bottom_right_x = ?, bounding_box_bottom_right_y = ?,
			bounding_box_width = ?, bounding_box_height = ?, video_frame = ?,
			filter_applied = ?, duration = ?, movement_speed = ?,
			movement_direction = ?, movement_pattern = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		sls.Timestamp, skeletalJSON, sls.Confidence,
		sls.BoundingBox.TopLeft.X, sls.BoundingBox.TopLeft.Y,
		sls.BoundingBox.BottomRight.X, sls.BoundingBox.BottomRight.Y,
		sls.BoundingBox.Width, sls.BoundingBox.Height, sls.VideoFrame,
		filtersJSON, sls.Duration, sls.Movement.Speed,
		sls.Movement.Direction, sls.Movement.Pattern, sls.ID,
	)

	return err
}

// Delete deletes an SLS detection
func (r *SQLiteSLSRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM sls_detections WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByConfidenceRange retrieves SLS detections by minimum confidence
func (r *SQLiteSLSRepository) GetByConfidenceRange(ctx context.Context, minConfidence float64) ([]*domain.SLSDetection, error) {
	query := `
		SELECT id, session_id, timestamp, skeletal_points, confidence,
			bounding_box_top_left_x, bounding_box_top_left_y,
			bounding_box_bottom_right_x, bounding_box_bottom_right_y,
			bounding_box_width, bounding_box_height, video_frame, filter_applied,
			duration, movement_speed, movement_direction, movement_pattern, created_at
		FROM sls_detections WHERE confidence >= ? ORDER BY confidence DESC`

	rows, err := r.db.QueryContext(ctx, query, minConfidence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slsDetections []*domain.SLSDetection
	for rows.Next() {
		var sls domain.SLSDetection
		var skeletalJSON, filtersJSON string

		err := rows.Scan(
			&sls.ID, &sls.SessionID, &sls.Timestamp, &skeletalJSON, &sls.Confidence,
			&sls.BoundingBox.TopLeft.X, &sls.BoundingBox.TopLeft.Y,
			&sls.BoundingBox.BottomRight.X, &sls.BoundingBox.BottomRight.Y,
			&sls.BoundingBox.Width, &sls.BoundingBox.Height, &sls.VideoFrame, &filtersJSON,
			&sls.Duration, &sls.Movement.Speed, &sls.Movement.Direction, &sls.Movement.Pattern,
			&sls.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(skeletalJSON), &sls.SkeletalPoints)
		json.Unmarshal([]byte(filtersJSON), &sls.FilterApplied)

		slsDetections = append(slsDetections, &sls)
	}

	return slsDetections, rows.Err()
}

// GetByDuration retrieves SLS detections by minimum duration
func (r *SQLiteSLSRepository) GetByDuration(ctx context.Context, minDuration float64) ([]*domain.SLSDetection, error) {
	query := `
		SELECT id, session_id, timestamp, skeletal_points, confidence,
			bounding_box_top_left_x, bounding_box_top_left_y,
			bounding_box_bottom_right_x, bounding_box_bottom_right_y,
			bounding_box_width, bounding_box_height, video_frame, filter_applied,
			duration, movement_speed, movement_direction, movement_pattern, created_at
		FROM sls_detections WHERE duration >= ? ORDER BY duration DESC`

	rows, err := r.db.QueryContext(ctx, query, minDuration)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slsDetections []*domain.SLSDetection
	for rows.Next() {
		var sls domain.SLSDetection
		var skeletalJSON, filtersJSON string

		err := rows.Scan(
			&sls.ID, &sls.SessionID, &sls.Timestamp, &skeletalJSON, &sls.Confidence,
			&sls.BoundingBox.TopLeft.X, &sls.BoundingBox.TopLeft.Y,
			&sls.BoundingBox.BottomRight.X, &sls.BoundingBox.BottomRight.Y,
			&sls.BoundingBox.Width, &sls.BoundingBox.Height, &sls.VideoFrame, &filtersJSON,
			&sls.Duration, &sls.Movement.Speed, &sls.Movement.Direction, &sls.Movement.Pattern,
			&sls.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(skeletalJSON), &sls.SkeletalPoints)
		json.Unmarshal([]byte(filtersJSON), &sls.FilterApplied)

		slsDetections = append(slsDetections, &sls)
	}

	return slsDetections, rows.Err()
}

// SQLiteInteractionRepository implements InteractionRepository using SQLite
type SQLiteInteractionRepository struct {
	db *sql.DB
}

// NewSQLiteInteractionRepository creates a new SQLite interaction repository
func NewSQLiteInteractionRepository(db *sql.DB) *SQLiteInteractionRepository {
	return &SQLiteInteractionRepository{db: db}
}

// Create creates a new user interaction
func (r *SQLiteInteractionRepository) Create(ctx context.Context, interaction *domain.UserInteraction) error {
	var randType, randResult, randRange sql.NullString
	if interaction.RandomizerResult != nil {
		randType = sql.NullString{String: interaction.RandomizerResult.Type, Valid: true}
		randRange = sql.NullString{String: interaction.RandomizerResult.Range, Valid: true}
		if resultBytes, err := json.Marshal(interaction.RandomizerResult.Result); err == nil {
			randResult = sql.NullString{String: string(resultBytes), Valid: true}
		}
	}

	query := `
		INSERT INTO user_interactions (
			id, session_id, timestamp, type, content, audio_path, response, response_time,
			randomizer_type, randomizer_result, randomizer_range, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		interaction.ID, interaction.SessionID, interaction.Timestamp, interaction.Type,
		interaction.Content, interaction.AudioPath, interaction.Response, interaction.ResponseTime,
		randType, randResult, randRange, interaction.CreatedAt,
	)

	return err
}

// GetByID retrieves a user interaction by ID
func (r *SQLiteInteractionRepository) GetByID(ctx context.Context, id string) (*domain.UserInteraction, error) {
	query := `
		SELECT id, session_id, timestamp, type, content, audio_path, response, response_time,
			randomizer_type, randomizer_result, randomizer_range, created_at
		FROM user_interactions WHERE id = ?`

	var interaction domain.UserInteraction
	var audioPath, response sql.NullString
	var responseTime sql.NullFloat64
	var randType, randResult, randRange sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&interaction.ID, &interaction.SessionID, &interaction.Timestamp, &interaction.Type,
		&interaction.Content, &audioPath, &response, &responseTime,
		&randType, &randResult, &randRange, &interaction.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if audioPath.Valid {
		interaction.AudioPath = audioPath.String
	}
	if response.Valid {
		interaction.Response = response.String
	}
	if responseTime.Valid {
		interaction.ResponseTime = responseTime.Float64
	}

	if randType.Valid && randResult.Valid && randRange.Valid {
		interaction.RandomizerResult = &domain.RandomizerResult{
			Type:  randType.String,
			Range: randRange.String,
		}
		json.Unmarshal([]byte(randResult.String), &interaction.RandomizerResult.Result)
	}

	return &interaction, nil
}

// GetBySessionID retrieves user interactions by session ID
func (r *SQLiteInteractionRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.UserInteraction, error) {
	query := `
		SELECT id, session_id, timestamp, type, content, audio_path, response, response_time,
			randomizer_type, randomizer_result, randomizer_range, created_at
		FROM user_interactions WHERE session_id = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interactions []*domain.UserInteraction
	for rows.Next() {
		var interaction domain.UserInteraction
		var audioPath, response sql.NullString
		var responseTime sql.NullFloat64
		var randType, randResult, randRange sql.NullString

		err := rows.Scan(
			&interaction.ID, &interaction.SessionID, &interaction.Timestamp, &interaction.Type,
			&interaction.Content, &audioPath, &response, &responseTime,
			&randType, &randResult, &randRange, &interaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if audioPath.Valid {
			interaction.AudioPath = audioPath.String
		}
		if response.Valid {
			interaction.Response = response.String
		}
		if responseTime.Valid {
			interaction.ResponseTime = responseTime.Float64
		}

		if randType.Valid && randResult.Valid && randRange.Valid {
			interaction.RandomizerResult = &domain.RandomizerResult{
				Type:  randType.String,
				Range: randRange.String,
			}
			json.Unmarshal([]byte(randResult.String), &interaction.RandomizerResult.Result)
		}

		interactions = append(interactions, &interaction)
	}

	return interactions, rows.Err()
}

// Update updates a user interaction
func (r *SQLiteInteractionRepository) Update(ctx context.Context, interaction *domain.UserInteraction) error {
	var randType, randResult, randRange sql.NullString
	if interaction.RandomizerResult != nil {
		randType = sql.NullString{String: interaction.RandomizerResult.Type, Valid: true}
		randRange = sql.NullString{String: interaction.RandomizerResult.Range, Valid: true}
		if resultBytes, err := json.Marshal(interaction.RandomizerResult.Result); err == nil {
			randResult = sql.NullString{String: string(resultBytes), Valid: true}
		}
	}

	query := `
		UPDATE user_interactions SET
			timestamp = ?, type = ?, content = ?, audio_path = ?, response = ?,
			response_time = ?, randomizer_type = ?, randomizer_result = ?, randomizer_range = ?
		WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		interaction.Timestamp, interaction.Type, interaction.Content, interaction.AudioPath,
		interaction.Response, interaction.ResponseTime, randType, randResult, randRange,
		interaction.ID,
	)

	return err
}

// Delete deletes a user interaction
func (r *SQLiteInteractionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM user_interactions WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByType retrieves user interactions by type
func (r *SQLiteInteractionRepository) GetByType(ctx context.Context, interactionType domain.InteractionType) ([]*domain.UserInteraction, error) {
	query := `
		SELECT id, session_id, timestamp, type, content, audio_path, response, response_time,
			randomizer_type, randomizer_result, randomizer_range, created_at
		FROM user_interactions WHERE type = ? ORDER BY timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, interactionType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interactions []*domain.UserInteraction
	for rows.Next() {
		var interaction domain.UserInteraction
		var audioPath, response sql.NullString
		var responseTime sql.NullFloat64
		var randType, randResult, randRange sql.NullString

		err := rows.Scan(
			&interaction.ID, &interaction.SessionID, &interaction.Timestamp, &interaction.Type,
			&interaction.Content, &audioPath, &response, &responseTime,
			&randType, &randResult, &randRange, &interaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if audioPath.Valid {
			interaction.AudioPath = audioPath.String
		}
		if response.Valid {
			interaction.Response = response.String
		}
		if responseTime.Valid {
			interaction.ResponseTime = responseTime.Float64
		}

		if randType.Valid && randResult.Valid && randRange.Valid {
			interaction.RandomizerResult = &domain.RandomizerResult{
				Type:  randType.String,
				Range: randRange.String,
			}
			json.Unmarshal([]byte(randResult.String), &interaction.RandomizerResult.Result)
		}

		interactions = append(interactions, &interaction)
	}

	return interactions, rows.Err()
}

// SQLiteFileRepository implements FileRepository using SQLite and file system
type SQLiteFileRepository struct {
	db       *sql.DB
	basePath string
}

// NewSQLiteFileRepository creates a new SQLite file repository
func NewSQLiteFileRepository(db *sql.DB, basePath string) *SQLiteFileRepository {
	return &SQLiteFileRepository{
		db:       db,
		basePath: basePath,
	}
}

// SaveFile saves a file to the filesystem and records metadata
func (r *SQLiteFileRepository) SaveFile(ctx context.Context, path string, data []byte) error {
	fullPath := filepath.Join(r.basePath, path)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// Write file
	if err := ioutil.WriteFile(fullPath, data, 0644); err != nil {
		return err
	}

	// Record metadata in database
	query := `
		INSERT INTO files (id, session_id, file_path, file_type, file_size, mime_type, created_at)
		VALUES (?, '', ?, 'user_file', ?, '', datetime('now'))`

	id := strconv.FormatInt(ctx.Value("timestamp").(int64), 10)
	_, err := r.db.ExecContext(ctx, query, id, path, len(data))
	return err
}

// GetFile retrieves a file from the filesystem
func (r *SQLiteFileRepository) GetFile(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(r.basePath, path)
	return ioutil.ReadFile(fullPath)
}

// DeleteFile deletes a file from filesystem and database
func (r *SQLiteFileRepository) DeleteFile(ctx context.Context, path string) error {
	fullPath := filepath.Join(r.basePath, path)

	// Delete from filesystem
	if err := os.Remove(fullPath); err != nil {
		return err
	}

	// Delete from database
	query := `DELETE FROM files WHERE file_path = ?`
	_, err := r.db.ExecContext(ctx, query, path)
	return err
}

// FileExists checks if a file exists
func (r *SQLiteFileRepository) FileExists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(r.basePath, path)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// GetFileSize returns the size of a file
func (r *SQLiteFileRepository) GetFileSize(ctx context.Context, path string) (int64, error) {
	fullPath := filepath.Join(r.basePath, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ListFiles lists all files in a directory
func (r *SQLiteFileRepository) ListFiles(ctx context.Context, directory string) ([]string, error) {
	fullPath := filepath.Join(r.basePath, directory)

	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	return fileNames, nil
}
