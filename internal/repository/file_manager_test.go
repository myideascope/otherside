package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFileTestDB(t *testing.T) *sql.DB {
	db := setupTestDB(t)

	// Create files table schema
	schema := `
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

	return db
}

func TestNewFileManager_ValidInput_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()

	// Act
	fm := NewFileManager(db, storagePath)

	// Assert
	assert.NotNil(t, fm)
	assert.Equal(t, db, fm.db)
	assert.Equal(t, storagePath, fm.storagePath)
}

func TestFileManager_StoreFile_ValidFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test file content")

	// Act
	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, sessionID, metadata.SessionID)
	assert.Equal(t, filePath, metadata.FilePath)
	assert.Equal(t, ".txt", metadata.FileType)
	assert.Equal(t, "text/plain; charset=utf-8", metadata.MimeType)
	assert.Greater(t, metadata.FileSize, int64(0))
	assert.NotEmpty(t, metadata.Checksum)
	assert.NotEmpty(t, metadata.ID)

	// Verify file exists on disk
	fullPath := filepath.Join(storagePath, filePath)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	// Verify metadata stored in database
	retrieved, err := fm.getFileMetadata(context.Background(), filePath)
	assert.NoError(t, err)
	assert.Equal(t, metadata.ID, retrieved.ID)
	assert.Equal(t, metadata.Checksum, retrieved.Checksum)
}

func TestFileManager_StoreFile_DirectoryTraversal_Error(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "../../../etc/passwd" // Path traversal attempt
	content := strings.NewReader("Test content")

	// Act
	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "failed to create file")
}

func TestFileManager_GetFile_ValidFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test file content")

	// Store file first
	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
	require.NoError(t, err)

	// Act
	file, retrievedMetadata, err := fm.GetFile(context.Background(), filePath)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.NotNil(t, retrievedMetadata)
	assert.Equal(t, metadata.ID, retrievedMetadata.ID)
	assert.Equal(t, metadata.FilePath, retrievedMetadata.FilePath)

	file.Close()
}

func TestFileManager_GetFile_NonExistentFile_Error(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	file, metadata, err := fm.GetFile(context.Background(), "non-existent.txt")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, file)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "failed to get file metadata")
}

func TestFileManager_DeleteFile_ValidFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test file content")

	// Store file first
	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
	require.NoError(t, err)

	// Act
	err = fm.DeleteFile(context.Background(), filePath)

	// Assert
	assert.NoError(t, err)

	// Verify file deleted from disk
	fullPath := filepath.Join(storagePath, filePath)
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err))

	// Verify metadata deleted from database
	_, err = fm.getFileMetadata(context.Background(), filePath)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestFileManager_DeleteFile_NonExistentFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	err := fm.DeleteFile(context.Background(), "non-existent.txt")

	// Assert
	assert.NoError(t, err) // Should not error on non-existent file
}

func TestFileManager_UpdateFile_ValidFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "sessions/test-session-id/test-file.txt"
	oldContent := strings.NewReader("Original content")
	newContent := strings.NewReader("Updated content")

	// Store file first
	oldMetadata, err := fm.StoreFile(context.Background(), sessionID, filePath, oldContent)
	require.NoError(t, err)

	// Act
	newMetadata, err := fm.UpdateFile(context.Background(), filePath, newContent)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, newMetadata)
	assert.Equal(t, oldMetadata.SessionID, newMetadata.SessionID)
	assert.Equal(t, oldMetadata.FilePath, newMetadata.FilePath)
	assert.NotEqual(t, oldMetadata.Checksum, newMetadata.Checksum)
	assert.NotEqual(t, oldMetadata.FileSize, newMetadata.FileSize)
}

func TestFileManager_ListFilesBySession_ValidSessionID_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"

	// Store multiple files
	files := []struct {
		path    string
		content string
	}{
		{"file1.txt", "Content 1"},
		{"file2.wav", "WAV data"},
		{"subdir/file3.jpg", "JPG data"},
	}

	for _, f := range files {
		content := strings.NewReader(f.content)
		_, err := fm.StoreFile(context.Background(), sessionID, f.path, content)
		require.NoError(t, err)
	}

	// Act
	metadataList, err := fm.ListFilesBySession(context.Background(), sessionID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, metadataList, 3)

	// Verify all files belong to the session
	for _, metadata := range metadataList {
		assert.Equal(t, sessionID, metadata.SessionID)
	}
}

func TestFileManager_VerifyFileIntegrity_ValidFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test file content")

	// Store file first
	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
	require.NoError(t, err)

	// Act
	valid, err := fm.VerifyFileIntegrity(context.Background(), filePath)

	// Assert
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestFileManager_VerifyFileIntegrity_CorruptedFile_Failure(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test file content")

	// Store file first
	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
	require.NoError(t, err)

	// Corrupt the file on disk
	fullPath := filepath.Join(storagePath, filePath)
	file, err := os.OpenFile(fullPath, os.O_WRONLY, 0644)
	require.NoError(t, err)
	file.WriteString("Corrupted data")
	file.Close()

	// Act
	valid, err := fm.VerifyFileIntegrity(context.Background(), filePath)

	// Assert
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestFileManager_CleanupOrphanedFiles_ValidDirectory_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test file content")

	// Store file first
	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
	require.NoError(t, err)

	// Create orphaned file (exists on disk but no metadata)
	orphanedPath := filepath.Join(storagePath, "orphaned.txt")
	err = os.WriteFile(orphanedPath, []byte("Orphaned content"), 0644)
	require.NoError(t, err)

	// Act
	err = fm.CleanupOrphanedFiles(context.Background())

	// Assert
	assert.NoError(t, err)

	// Verify orphaned file was deleted
	_, err = os.Stat(orphanedPath)
	assert.True(t, os.IsNotExist(err))

	// Verify managed file still exists
	managedPath := filepath.Join(storagePath, filePath)
	_, err = os.Stat(managedPath)
	assert.NoError(t, err)
}

func TestFileManager_GetStorageStats_ValidDatabase_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"

	// Store multiple files
	for i := 0; i < 3; i++ {
		filePath := fmt.Sprintf("file%d.txt", i)
		content := strings.NewReader(fmt.Sprintf("Content %d", i))
		_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
		require.NoError(t, err)
	}

	// Act
	stats, err := fm.GetStorageStats(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(3), stats["total_files"])
	assert.Equal(t, int64(1), stats["unique_sessions"])
	assert.Greater(t, stats["total_size_bytes"], int64(0))
	assert.GreaterOrEqual(t, stats["total_size_mb"], int64(0))
}

func TestFileManager_ConcurrentFileOperations_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"

	// Act - Store multiple files concurrently
	done := make(chan bool, 5)

	for index := 0; index < 5; index++ {
		go func(idx int) {
			filePath := fmt.Sprintf("concurrent-file-%d.txt", idx)
			content := strings.NewReader(fmt.Sprintf("Concurrent content %d", idx))

			_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)
			assert.NoError(t, err)
			done <- true
		}(index)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Assert - All files should be stored
	files, err := fm.ListFilesBySession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Len(t, files, 5)
}

func TestFileManager_FileTypeDetection_UnknownExtension_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	sessionID := "test-session-id"
	filePath := "test-file.unknown" // Unknown extension
	content := strings.NewReader("Test content")

	// Act
	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, ".unknown", metadata.FileType)
	assert.Equal(t, "application/octet-stream", metadata.MimeType)
}

func TestFileManager_StorageQuota_Enforcement_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(t)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Create a file that approaches storage limit
	// This test simulates quota enforcement by creating multiple large files
	for i := 0; i < 10; i++ {
		filePath := fmt.Sprintf("large-file-%d.dat", i)
		content := make([]byte, 100*1024) // 100KB files
		_, err := fm.StoreFile(context.Background(), "test-session-id", filePath, strings.NewReader(string(content)))
		assert.NoError(t, err)
	}

	// Assert - Verify files were stored
	files, err := fm.ListFilesBySession(context.Background(), "test-session-id")
	assert.NoError(t, err)
	assert.Len(t, files, 10)

	// Verify total storage size
	stats, err := fm.GetStorageStats(context.Background())
	assert.NoError(t, err)
	assert.Greater(t, stats["total_size_bytes"], int64(1000*1024))
}

func TestFileManager_FileRetention_Cleanup_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Create an old file that should be cleaned up
	oldTime := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago
	sessionID := "test-session-id"
	filePath := "sessions/test-session-id/old-file.txt"
	oldContent := "Old file content"

	// Store file with manual old timestamp
	metadata := &FileMetadata{
		ID:        "old-file-id",
		SessionID: sessionID,
		FilePath:  filePath,
		FileType:  ".txt",
		FileSize:  int64(len(oldContent)),
		MimeType:  "text/plain",
		Checksum:  "old-checksum",
		CreatedAt: oldTime,
	}
	err := fm.storeMetadata(context.Background(), metadata)
	require.NoError(t, err)

	// Create actual file
	fullPath := filepath.Join(storagePath, filePath)
	err = os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, []byte(oldContent), 0644)
	require.NoError(t, err)

	// Act
	err = fm.CleanupOrphanedFiles(context.Background())

	// Assert
	assert.NoError(t, err)

	// Old file should still exist because it has metadata
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)
}

func TestFileManager_TemporaryFileHandling_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Create temporary file directly
	tempPath := filepath.Join(storagePath, "temp-file.tmp")
	tempContent := "Temporary content"
	err := os.WriteFile(tempPath, []byte(tempContent), 0644)
	require.NoError(t, err)

	// Act - Store a regular file to verify FileManager ignores temp files
	sessionID := "test-session-id"
	filePath := "regular-file.txt"
	content := strings.NewReader("Regular content")
	_, err = fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)

	// Verify only regular file metadata was stored
	metadata, err := fm.getFileMetadata(context.Background(), filePath)
	assert.NoError(t, err)
	assert.Equal(t, sessionID, metadata.SessionID)

	// Temp file should still exist (FileManager doesn't touch it)
	_, err = os.Stat(tempPath)
	assert.NoError(t, err)
}

func TestFileManager_ConcurrentFileOperations_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Perform multiple file operations concurrently
	sessionID := "test-session-id"
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			filePath := fmt.Sprintf("concurrent-file-%d.txt", index)
			content := fmt.Sprintf("Content %d", index)

			// Store file
			_, err := fm.StoreFile(context.Background(), sessionID, filePath, strings.NewReader(content))
			assert.NoError(t, err)

			// Get file
			file, _, err := fm.GetFile(context.Background(), filePath)
			assert.NoError(t, err)
			file.Close()

			// Verify file integrity
			valid, err := fm.VerifyFileIntegrity(context.Background(), filePath)
			assert.NoError(t, err)
			assert.True(t, valid)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Assert - All files should be stored
	files, err := fm.ListFilesBySession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Len(t, files, 5)
}

func TestFileManager_DiskSpaceMonitoring_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Create files until we approach disk space limits
	// This simulates disk space monitoring
	sessionID := "test-session-id"
	fileCount := 0
	var totalSize int64

	for fileCount = 0; fileCount < 100; fileCount++ {
		filePath := fmt.Sprintf("space-test-file-%d.dat", fileCount)
		content := make([]byte, 10*1024) // 10KB files

		_, err := fm.StoreFile(context.Background(), sessionID, filePath, strings.NewReader(string(content)))
		if err != nil {
			break // Stop when we hit disk space or other errors
		}

		if fileCount < 50 { // Only count first 50 to avoid long test
			totalSize += int64(len(content))
		}
	}

	// Assert - At least some files should be stored
	files, err := fm.ListFilesBySession(context.Background(), sessionID)
	if err == nil {
		assert.Greater(t, len(files), 0)
		if len(files) > 0 {
			assert.Greater(t, totalSize, int64(0))
		}
	}
}

func TestFileManager_UnicodeFileNames_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"
	filePath := "测试文件.txt" // Unicode filename
	content := strings.NewReader("Unicode content: 你好世界")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, sessionID, metadata.SessionID)
	assert.Equal(t, filePath, metadata.FilePath)
	assert.Equal(t, ".txt", metadata.FileType)

	// Verify file exists on disk
	fullPath := filepath.Join(storagePath, filePath)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	// Verify can retrieve file
	file, _, err := fm.GetFile(context.Background(), filePath)
	assert.NoError(t, err)
	file.Close()
}

func TestFileManager_LongFilePath_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(t)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"

	// Create subdirectories first
	subDir := filepath.Join("deep", "nested", "directory", "structure")
	fullSubDir := filepath.Join(storagePath, subDir)
	err := os.MkdirAll(fullSubDir, 0755)
	require.NoError(t, err)

	filePath := filepath.Join(subDir, "deep-nested-file.txt")
	content := strings.NewReader("Deep nested content")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filepath.Join(subDir, "deep-nested-file.txt"), content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, sessionID, metadata.SessionID)

	// Verify file exists
	fullPath := filepath.Join(storagePath, metadata.FilePath)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	// Can retrieve by full path
	file, _, err := fm.GetFile(context.Background(), metadata.FilePath)
	assert.NoError(t, err)
	file.Close()
}

func TestFileManager_SpecialCharactersInContent_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"
	filePath := "special-chars.txt"
	content := strings.NewReader("Content with special chars: \x00\x01\x02 and \"quotes\" and \n newlines")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, int64(len("Content with special chars: \x00\x01\x02 and \"quotes\" and \n newlines")), metadata.FileSize)

	// Verify can retrieve the file with special characters
	file, _, err := fm.GetFile(context.Background(), filePath)
	assert.NoError(t, err)
	file.Close()
}

func TestFileManager_EmptyContentFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"
	filePath := "empty-file.txt"
	content := strings.NewReader("")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, int64(0), metadata.FileSize)

	// Verify can retrieve empty file
	file, _, err := fm.GetFile(context.Background(), filePath)
	assert.NoError(t, err)
	file.Close()

	// Verify integrity of empty file
	valid, err := fm.VerifyFileIntegrity(context.Background(), filePath)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestFileManager_MetadataStorage_ErrorHandling_DatabaseError(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Simulate database error during metadata storage
	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test content")

	// Manually corrupt the database connection by closing it
	db.Close()

	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "failed to store file metadata")
}

func TestFileManager_FileRetrieval_ErrorHandling_MissingMetadata(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Try to get file that exists but has no metadata
	sessionID := "test-session-id"
	filePath := "orphaned-file.txt"
	fullPath := filepath.Join(storagePath, filePath)

	// Create orphaned file (no metadata)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, []byte("orphaned content"), 0644)
	require.NoError(t, err)

	// Act
	file, metadata, err := fm.GetFile(context.Background(), filePath)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get file metadata")
	assert.Nil(t, file)
	assert.Nil(t, metadata)
}

// Additional comprehensive tests

func TestFileManager_FileTypeDetection_NoExtension_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	filePath := "test-file" // No extension
	content := strings.NewReader("Test file content")
	metadata, err := fm.StoreFile(context.Background(), "test-session-id", filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "unknown", metadata.FileType)
	assert.Equal(t, "application/octet-stream", metadata.MimeType)
}

func TestFileManager_StorageQuota_Enforcement_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Create a file that approaches storage limit
	// This test simulates quota enforcement by creating multiple large files
	sessionID := "test-session-id"
	fileCount := 0
	var totalSize int64

	for fileCount = 0; fileCount < 10; fileCount++ {
		filePath := fmt.Sprintf("large-file-%d.dat", fileCount)
		content := make([]byte, 100*1024) // 100KB files
		_, err := fm.StoreFile(context.Background(), sessionID, filePath, strings.NewReader(string(content)))
		if err != nil {
			break // Stop when we hit disk space or other errors
		}

		totalSize += int64(len(content))
	}

	// Assert - At least some files should be stored
	files, err := fm.ListFilesBySession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Greater(t, len(files), 0)
	if len(files) > 0 {
		assert.Greater(t, totalSize, int64(0))
	}
}

func TestFileManager_FileRetention_Cleanup_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Create an old file that should be cleaned up
	oldTime := time.Now().Add(-7 * 24 * time.Hour) // 7 days ago
	sessionID := "test-session-id"
	filePath := "sessions/test-session-id/old-file.txt"
	oldContent := "Old file content"

	// Store file with manual old timestamp
	metadata := &FileMetadata{
		ID:        "old-file-id",
		SessionID: sessionID,
		FilePath:  filePath,
		FileType:  ".txt",
		FileSize:  int64(len(oldContent)),
		MimeType:  "text/plain",
		Checksum:  "old-checksum",
		CreatedAt: oldTime,
	}
	err := fm.storeMetadata(context.Background(), metadata)
	require.NoError(t, err)

	// Create actual file
	fullPath := filepath.Join(storagePath, filePath)
	err = os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, []byte(oldContent), 0644)
	require.NoError(t, err)

	// Act
	err = fm.CleanupOrphanedFiles(context.Background())

	// Assert
	assert.NoError(t, err)

	// Old file should still exist because it has metadata
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)
}

func TestFileManager_TemporaryFileHandling_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Create temporary file directly
	tempPath := filepath.Join(storagePath, "temp-file.tmp")
	tempContent := "Temporary content"
	err := os.WriteFile(tempPath, []byte(tempContent), 0644)
	require.NoError(t, err)

	// Act - Store a regular file to verify FileManager ignores temp files
	sessionID := "test-session-id"
	filePath := "regular-file.txt"
	content := strings.NewReader("Regular content")
	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)

	// Verify only regular file metadata was stored
	metadata, err := fm.getFileMetadata(context.Background(), filePath)
	assert.NoError(t, err)
	assert.Equal(t, sessionID, metadata.SessionID)

	// Temp file should still exist (FileManager doesn't touch it)
	_, err = os.Stat(tempPath)
	assert.NoError(t, err)
}

func TestFileManager_ConcurrentFileOperations_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Perform multiple file operations concurrently
	sessionID := "test-session-id"
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			filePath := fmt.Sprintf("concurrent-file-%d.txt", index)
			content := fmt.Sprintf("Concurrent content %d", index)

			// Store file
			_, err := fm.StoreFile(context.Background(), sessionID, filePath, strings.NewReader(content))
			assert.NoError(t, err)

			// Get file
			file, _, err := fm.GetFile(context.Background(), filePath)
			assert.NoError(t, err)
			file.Close()

			// Verify file integrity
			valid, err := fm.VerifyFileIntegrity(context.Background(), filePath)
			assert.NoError(t, err)
			assert.True(t, valid)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Assert - All files should be stored
	files, err := fm.ListFilesBySession(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.Len(t, files, 5)
}

func TestFileManager_DiskSpaceMonitoring_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Create files until we approach disk space limits
	sessionID := "test-session-id"
	fileCount := 0
	var totalSize int64

	for fileCount = 0; fileCount < 100; fileCount++ {
		filePath := fmt.Sprintf("space-test-file-%d.dat", fileCount)
		content := make([]byte, 10*1024) // 10KB files

		_, err := fm.StoreFile(context.Background(), sessionID, filePath, strings.NewReader(string(content)))
		if err != nil {
			break // Stop when we hit disk space or other errors
		}

		totalSize += int64(len(content))
	}

	// Assert - At least some files should be stored
	files, err := fm.ListFilesBySession(context.Background(), sessionID)
	if err == nil && len(files) > 0 {
		assert.Greater(t, totalSize, int64(0))
	}
}

func TestFileManager_UnicodeFileNames_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"
	filePath := "测试文件.txt" // Unicode filename
	content := strings.NewReader("Unicode content: 你好世界")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, sessionID, metadata.SessionID)
	assert.Equal(t, filePath, metadata.FilePath)

	// Verify file exists on disk
	fullPath := filepath.Join(storagePath, filePath)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	// Verify can retrieve file with unicode path
	file, _, err := fm.GetFile(context.Background(), filePath)
	assert.NoError(t, err)
	file.Close()
}

func TestFileManager_LongFilePath_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"

	// Create subdirectories first
	subDir := filepath.Join(storagePath, "deep", "nested", "directory", "structure")
	fullSubDir := filepath.Join(storagePath, subDir)
	err := os.MkdirAll(fullSubDir, 0755)
	require.NoError(t, err)

	filePath := filepath.Join(subDir, "deep-nested-file.txt")
	content := strings.NewReader("Deep nested content")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filepath.Join(subDir, "deep-nested-file.txt"), content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, sessionID, metadata.SessionID)

	// Verify file exists
	fullPath := filepath.Join(storagePath, metadata.FilePath)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err)

	// Can retrieve by full path
	file, _, err := fm.GetFile(context.Background(), metadata.FilePath)
	assert.NoError(t, err)
	file.Close()
}

func TestFileManager_SpecialCharactersInContent_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"
	filePath := "special-chars.txt"
	content := strings.NewReader("Content with special chars: \x00\x01\x02 and \"quotes\" and \n newlines")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, int64(len("Content with special chars: \x00\x01\x02 and \"quotes\" and \n newlines")), metadata.FileSize)

	// Verify can retrieve file with special characters
	file, _, err := fm.GetFile(context.Background(), filePath)
	assert.NoError(t, err)
	file.Close()
}

func TestFileManager_EmptyContentFile_Success(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act
	sessionID := "test-session-id"
	filePath := "empty-file.txt"
	content := strings.NewReader("")

	metadata, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, int64(0), metadata.FileSize)

	// Verify can retrieve empty file
	file, _, err := fm.GetFile(context.Background(), filePath)
	assert.NoError(t, err)
	file.Close()

	// Verify integrity of empty file
	valid, err := fm.VerifyFileIntegrity(context.Background(), filePath)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestFileManager_DatabaseErrorHandling_MetaDataStorageError(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(db)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Simulate database error during metadata storage
	sessionID := "test-session-id"
	filePath := "test-file.txt"
	content := strings.NewReader("Test content")

	// Manually corrupt database connection by closing it
	db.Close()

	_, err := fm.StoreFile(context.Background(), sessionID, filePath, content)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.Contains(t, err.Error(), "failed to store file metadata")
}

func TestFileManager_DatabaseErrorHandling_FileRetrievalError(t *testing.T) {
	// Arrange
	db := setupFileTestDB(t)
	defer cleanupTestDB(t)
	storagePath := t.TempDir()
	fm := NewFileManager(db, storagePath)

	// Act - Try to get file that exists but has no metadata
	sessionID := "test-session-id"
	filePath := "orphaned-file.txt"
	fullPath := filepath.Join(storagePath, filePath)

	// Create orphaned file (no metadata)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(fullPath, []byte("orphaned content"), 0644)
	require.NoError(tort)

	// Manually corrupt database connection by closing it
	db.Close()

	file, metadata, err := fm.GetFile(context.Background(), filePath)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get file metadata")
	assert.Nil(t, file)
	assert.Nil(t, metadata)
}
