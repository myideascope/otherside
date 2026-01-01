package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// CleanupManager manages cleanup operations for the database and file system
type CleanupManager struct {
	db          *sql.DB
	fileManager *FileManager
	sessionRepo *SQLiteSessionRepository
}

// CleanupConfig holds configuration for cleanup operations
type CleanupConfig struct {
	MaxSessionAge         time.Duration `json:"max_session_age"`
	MaxFileSizeBytes      int64         `json:"max_file_size_bytes"`
	MaxStorageSizeBytes   int64         `json:"max_storage_size_bytes"`
	EnableFileCleanup     bool          `json:"enable_file_cleanup"`
	EnableDatabaseCleanup bool          `json:"enable_database_cleanup"`
	EnableOrphanCleanup   bool          `json:"enable_orphan_cleanup"`
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(db *sql.DB, fileManager *FileManager) *CleanupManager {
	return &CleanupManager{
		db:          db,
		fileManager: fileManager,
		sessionRepo: NewSQLiteSessionRepository(db),
	}
}

// CleanupStats holds statistics about cleanup operations
type CleanupStats struct {
	SessionsCleaned  int                    `json:"sessions_cleaned"`
	FilesCleaned     int                    `json:"files_cleaned"`
	BytesFreed       int64                  `json:"bytes_freed"`
	OrphanFilesFound int                    `json:"orphan_files_found"`
	Duration         time.Duration          `json:"duration"`
	Errors           []string               `json:"errors"`
	StorageStats     map[string]interface{} `json:"storage_stats"`
}

// RunCleanup performs all cleanup operations
func (cm *CleanupManager) RunCleanup(ctx context.Context, config CleanupConfig) (*CleanupStats, error) {
	startTime := time.Now()
	stats := &CleanupStats{
		Errors:       make([]string, 0),
		StorageStats: make(map[string]interface{}),
	}

	log.Println("Starting cleanup operations...")

	// Get initial storage stats
	if err := cm.getStorageStats(ctx, stats); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to get initial storage stats: %v", err))
	}

	// Database cleanup
	if config.EnableDatabaseCleanup {
		if err := cm.cleanupOldSessions(ctx, config, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Database cleanup failed: %v", err))
		}
	}

	// File cleanup
	if config.EnableFileCleanup {
		if err := cm.cleanupLargeFiles(ctx, config, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("File cleanup failed: %v", err))
		}

		if err := cm.enforceStorageQuota(ctx, config, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Storage quota enforcement failed: %v", err))
		}
	}

	// Orphaned file cleanup
	if config.EnableOrphanCleanup {
		if err := cm.cleanupOrphanedFiles(ctx, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Orphan cleanup failed: %v", err))
		}
	}

	// Get final storage stats
	if err := cm.getStorageStats(ctx, stats); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to get final storage stats: %v", err))
	}

	stats.Duration = time.Since(startTime)

	// Log summary
	log.Printf("Cleanup completed in %v", stats.Duration)
	log.Printf("Sessions cleaned: %d, Files cleaned: %d, Bytes freed: %d",
		stats.SessionsCleaned, stats.FilesCleaned, stats.BytesFreed)

	if len(stats.Errors) > 0 {
		log.Printf("Cleanup completed with %d errors", len(stats.Errors))
		for _, err := range stats.Errors {
			log.Printf("Error: %s", err)
		}
	}

	return stats, nil
}

// cleanupOldSessions removes old completed/archived sessions
func (cm *CleanupManager) cleanupOldSessions(ctx context.Context, config CleanupConfig, stats *CleanupStats) error {
	if config.MaxSessionAge <= 0 {
		log.Println("Session age cleanup disabled")
		return nil
	}

	cutoffDate := time.Now().Add(-config.MaxSessionAge)
	log.Printf("Cleaning up sessions older than %v", cutoffDate)

	// Find old completed and archived sessions
	query := `
		SELECT id, COUNT(*) as file_count
		FROM sessions s
		LEFT JOIN files f ON s.id = f.session_id
		WHERE s.status IN ('complete', 'archived') 
		  AND s.created_at < ?
		GROUP BY s.id`

	rows, err := cm.db.QueryContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to query old sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sessionID string
		var fileCount int
		if err := rows.Scan(&sessionID, &fileCount); err != nil {
			log.Printf("Failed to scan session row: %v", err)
			continue
		}

		// Delete session (cascades to related records)
		if err := cm.sessionRepo.Delete(ctx, sessionID); err != nil {
			log.Printf("Failed to delete session %s: %v", sessionID, err)
			continue
		}

		stats.SessionsCleaned++
		stats.FilesCleaned += fileCount
	}

	return rows.Err()
}

// cleanupLargeFiles removes files exceeding maximum size
func (cm *CleanupManager) cleanupLargeFiles(ctx context.Context, config CleanupConfig, stats *CleanupStats) error {
	if config.MaxFileSizeBytes <= 0 {
		log.Println("Large file cleanup disabled")
		return nil
	}

	log.Printf("Cleaning up files larger than %d bytes", config.MaxFileSizeBytes)

	query := `SELECT file_path, file_size FROM files WHERE file_size > ?`
	rows, err := cm.db.QueryContext(ctx, query, config.MaxFileSizeBytes)
	if err != nil {
		return fmt.Errorf("failed to query large files: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filePath string
		var fileSize int64
		if err := rows.Scan(&filePath, &fileSize); err != nil {
			log.Printf("Failed to scan file row: %v", err)
			continue
		}

		// Delete file
		if err := cm.fileManager.DeleteFile(ctx, filePath); err != nil {
			log.Printf("Failed to delete large file %s: %v", filePath, err)
			continue
		}

		stats.FilesCleaned++
		stats.BytesFreed += fileSize
	}

	return rows.Err()
}

// enforceStorageQuota removes oldest files until storage quota is met
func (cm *CleanupManager) enforceStorageQuota(ctx context.Context, config CleanupConfig, stats *CleanupStats) error {
	if config.MaxStorageSizeBytes <= 0 {
		log.Println("Storage quota enforcement disabled")
		return nil
	}

	// Get current storage usage
	storageStats, err := cm.fileManager.GetStorageStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get storage stats: %w", err)
	}

	currentSize, ok := storageStats["total_size_bytes"].(int64)
	if !ok {
		return fmt.Errorf("invalid storage size type")
	}

	if currentSize <= config.MaxStorageSizeBytes {
		log.Printf("Storage usage %d bytes is within quota %d bytes", currentSize, config.MaxStorageSizeBytes)
		return nil
	}

	log.Printf("Storage usage %d bytes exceeds quota %d bytes, cleaning up oldest files",
		currentSize, config.MaxStorageSizeBytes)

	// Get oldest files, ordered by creation date
	query := `
		SELECT file_path, file_size 
		FROM files 
		ORDER BY created_at ASC`

	rows, err := cm.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query files for quota cleanup: %w", err)
	}
	defer rows.Close()

	targetSize := int64(float64(config.MaxStorageSizeBytes) * 0.9) // Clean up to 90% of quota

	for rows.Next() && currentSize > targetSize {
		var filePath string
		var fileSize int64
		if err := rows.Scan(&filePath, &fileSize); err != nil {
			log.Printf("Failed to scan file row: %v", err)
			continue
		}

		// Delete file
		if err := cm.fileManager.DeleteFile(ctx, filePath); err != nil {
			log.Printf("Failed to delete file %s for quota enforcement: %v", filePath, err)
			continue
		}

		stats.FilesCleaned++
		stats.BytesFreed += fileSize
		currentSize -= fileSize
	}

	return rows.Err()
}

// cleanupOrphanedFiles removes files without database metadata
func (cm *CleanupManager) cleanupOrphanedFiles(ctx context.Context, stats *CleanupStats) error {
	log.Println("Cleaning up orphaned files")

	if err := cm.fileManager.CleanupOrphanedFiles(ctx); err != nil {
		return fmt.Errorf("failed to cleanup orphaned files: %w", err)
	}

	// Note: CleanupOrphanedFiles doesn't return count, so we can't update stats directly
	// This could be enhanced to return the count of deleted files
	return nil
}

// getStorageStats captures current storage statistics
func (cm *CleanupManager) getStorageStats(ctx context.Context, stats *CleanupStats) error {
	storageStats, err := cm.fileManager.GetStorageStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get storage stats: %w", err)
	}

	stats.StorageStats = storageStats
	return nil
}

// GetCleanupPlan shows what would be cleaned up without actually cleaning
func (cm *CleanupManager) GetCleanupPlan(ctx context.Context, config CleanupConfig) (*CleanupStats, error) {
	stats := &CleanupStats{
		Errors:       make([]string, 0),
		StorageStats: make(map[string]interface{}),
	}

	// Get current stats
	if err := cm.getStorageStats(ctx, stats); err != nil {
		stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to get storage stats: %v", err))
		return stats, err
	}

	// Count what would be cleaned (without actually deleting)
	if config.EnableDatabaseCleanup && config.MaxSessionAge > 0 {
		cutoffDate := time.Now().Add(-config.MaxSessionAge)
		query := `SELECT COUNT(*) FROM sessions WHERE status IN ('complete', 'archived') AND created_at < ?`
		if err := cm.db.QueryRowContext(ctx, query, cutoffDate).Scan(&stats.SessionsCleaned); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to count old sessions: %v", err))
		}
	}

	if config.EnableFileCleanup && config.MaxFileSizeBytes > 0 {
		query := `SELECT COUNT(*), COALESCE(SUM(file_size), 0) FROM files WHERE file_size > ?`
		var fileSizeSum int64
		if err := cm.db.QueryRowContext(ctx, query, config.MaxFileSizeBytes).Scan(&stats.FilesCleaned, &fileSizeSum); err != nil {
			stats.Errors = append(stats.Errors, fmt.Sprintf("Failed to count large files: %v", err))
		} else {
			stats.BytesFreed += fileSizeSum
		}
	}

	return stats, nil
}
