package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FileMetadata represents file metadata stored in the database
type FileMetadata struct {
	ID        string    `json:"id" db:"id"`
	SessionID string    `json:"session_id" db:"session_id"`
	FilePath  string    `json:"file_path" db:"file_path"`
	FileType  string    `json:"file_type" db:"file_type"`
	FileSize  int64     `json:"file_size" db:"file_size"`
	MimeType  string    `json:"mime_type" db:"mime_type"`
	Checksum  string    `json:"checksum" db:"checksum"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FileManager manages file operations and metadata
type FileManager struct {
	db          *sql.DB
	storagePath string
}

// NewFileManager creates a new file manager
func NewFileManager(db *sql.DB, storagePath string) *FileManager {
	return &FileManager{
		db:          db,
		storagePath: storagePath,
	}
}

// StoreFile stores a file and records its metadata
func (fm *FileManager) StoreFile(ctx context.Context, sessionID, filePath string, file io.Reader) (*FileMetadata, error) {
	// Ensure storage directory exists
	if err := os.MkdirAll(fm.storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Determine full file path
	fullPath := filepath.Join(fm.storagePath, filePath)

	// Ensure directory for file exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create file directory: %w", err)
	}

	// Create the file
	fileHandle, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer fileHandle.Close()

	// Calculate checksum while writing
	hasher := sha256.New()
	multiWriter := io.MultiWriter(fileHandle, hasher)

	// Copy file data and calculate checksum
	size, err := io.Copy(multiWriter, file)
	if err != nil {
		os.Remove(fullPath) // Clean up on failure
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Generate metadata
	fileID := generateID()
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	fileType := filepath.Ext(filePath)
	if fileType == "" {
		fileType = "unknown"
	}
	mimeType := mime.TypeByExtension(fileType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	metadata := &FileMetadata{
		ID:        fileID,
		SessionID: sessionID,
		FilePath:  filePath,
		FileType:  fileType,
		FileSize:  size,
		MimeType:  mimeType,
		Checksum:  checksum,
		CreatedAt: time.Now(),
	}

	// Store metadata in database
	if err := fm.storeMetadata(ctx, metadata); err != nil {
		os.Remove(fullPath) // Clean up on failure
		return nil, fmt.Errorf("failed to store file metadata: %w", err)
	}

	log.Printf("Stored file: %s (size: %d bytes, checksum: %s)", filePath, size, checksum[:16]+"...")
	return metadata, nil
}

// GetFile retrieves a file and its metadata
func (fm *FileManager) GetFile(ctx context.Context, filePath string) (*os.File, *FileMetadata, error) {
	// Get metadata from database
	metadata, err := fm.getFileMetadata(ctx, filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Open file
	fullPath := filepath.Join(fm.storagePath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, metadata, nil
}

// DeleteFile removes a file and its metadata
func (fm *FileManager) DeleteFile(ctx context.Context, filePath string) error {
	// Get metadata first for logging
	metadata, err := fm.getFileMetadata(ctx, filePath)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Delete from database
	if err := fm.deleteMetadata(ctx, filePath); err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	// Delete physical file
	fullPath := filepath.Join(fm.storagePath, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if metadata != nil {
		log.Printf("Deleted file: %s (size: %d bytes)", filePath, metadata.FileSize)
	}

	return nil
}

// UpdateFile updates an existing file
func (fm *FileManager) UpdateFile(ctx context.Context, filePath string, newContent io.Reader) (*FileMetadata, error) {
	// Delete old file
	if err := fm.DeleteFile(ctx, filePath); err != nil {
		return nil, fmt.Errorf("failed to delete old file: %w", err)
	}

	// Get session ID from old metadata or extract from path
	sessionID := extractSessionIDFromPath(filePath)
	if sessionID == "" {
		return nil, fmt.Errorf("could not determine session ID from file path: %s", filePath)
	}

	// Store new file
	return fm.StoreFile(ctx, sessionID, filePath, newContent)
}

// ListFilesBySession returns all files for a session
func (fm *FileManager) ListFilesBySession(ctx context.Context, sessionID string) ([]*FileMetadata, error) {
	query := `
		SELECT id, session_id, file_path, file_type, file_size, 
		       mime_type, checksum, created_at 
		FROM files 
		WHERE session_id = ? 
		ORDER BY created_at DESC`

	rows, err := fm.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []*FileMetadata
	for rows.Next() {
		var metadata FileMetadata
		err := rows.Scan(
			&metadata.ID, &metadata.SessionID, &metadata.FilePath,
			&metadata.FileType, &metadata.FileSize, &metadata.MimeType,
			&metadata.Checksum, &metadata.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file metadata: %w", err)
		}
		files = append(files, &metadata)
	}

	return files, rows.Err()
}

// VerifyFileIntegrity checks if a file's checksum matches the stored value
func (fm *FileManager) VerifyFileIntegrity(ctx context.Context, filePath string) (bool, error) {
	// Get stored metadata
	metadata, err := fm.getFileMetadata(ctx, filePath)
	if err != nil {
		return false, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Open file
	fullPath := filepath.Join(fm.storagePath, filePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Calculate checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return false, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	calculatedChecksum := fmt.Sprintf("%x", hasher.Sum(nil))
	matches := calculatedChecksum == metadata.Checksum

	if !matches {
		log.Printf("File integrity check failed for %s: expected %s, got %s",
			filePath, metadata.Checksum[:16]+"...", calculatedChecksum[:16]+"...")
	}

	return matches, nil
}

// CleanupOrphanedFiles removes files without metadata
func (fm *FileManager) CleanupOrphanedFiles(ctx context.Context) error {
	log.Println("Cleaning up orphaned files...")

	// Get all files in storage directory
	err := filepath.Walk(fm.storagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(fm.storagePath, path)
		if err != nil {
			return err
		}

		// Check if metadata exists
		_, err = fm.getFileMetadata(ctx, relPath)
		if err == sql.ErrNoRows {
			// No metadata found, delete orphaned file
			if err := os.Remove(path); err != nil {
				log.Printf("Failed to delete orphaned file %s: %v", path, err)
			} else {
				log.Printf("Deleted orphaned file: %s", relPath)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk storage directory: %w", err)
	}

	log.Println("Orphaned file cleanup complete")
	return nil
}

// GetStorageStats returns storage usage statistics
func (fm *FileManager) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_files,
			COALESCE(SUM(file_size), 0) as total_size,
			COUNT(DISTINCT session_id) as unique_sessions
		FROM files`

	var totalFiles, totalSize, uniqueSessions int64
	err := fm.db.QueryRowContext(ctx, query).Scan(&totalFiles, &totalSize, &uniqueSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_files":      totalFiles,
		"total_size_bytes": totalSize,
		"total_size_mb":    totalSize / (1024 * 1024),
		"unique_sessions":  uniqueSessions,
	}

	return stats, nil
}

// Helper functions

func (fm *FileManager) storeMetadata(ctx context.Context, metadata *FileMetadata) error {
	query := `
		INSERT INTO files (id, session_id, file_path, file_type, file_size, mime_type, checksum, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := fm.db.ExecContext(ctx, query,
		metadata.ID, metadata.SessionID, metadata.FilePath,
		metadata.FileType, metadata.FileSize, metadata.MimeType,
		metadata.Checksum, metadata.CreatedAt)

	return err
}

func (fm *FileManager) getFileMetadata(ctx context.Context, filePath string) (*FileMetadata, error) {
	query := `
		SELECT id, session_id, file_path, file_type, file_size, 
		       mime_type, checksum, created_at 
		FROM files 
		WHERE file_path = ?`

	var metadata FileMetadata
	err := fm.db.QueryRowContext(ctx, query, filePath).Scan(
		&metadata.ID, &metadata.SessionID, &metadata.FilePath,
		&metadata.FileType, &metadata.FileSize, &metadata.MimeType,
		&metadata.Checksum, &metadata.CreatedAt)

	return &metadata, err
}

func (fm *FileManager) deleteMetadata(ctx context.Context, filePath string) error {
	query := `DELETE FROM files WHERE file_path = ?`
	_, err := fm.db.ExecContext(ctx, query, filePath)
	return err
}

func generateID() string {
	return fmt.Sprintf("file_%d", time.Now().UnixNano())
}

func extractSessionIDFromPath(filePath string) string {
	// Extract session ID from path like "sessions/session-id/recordings/audio.wav"
	parts := strings.Split(filepath.Clean(filePath), string(filepath.Separator))
	for i, part := range parts {
		if part == "sessions" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
