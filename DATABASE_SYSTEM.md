# Database Migration and Persistence System

This document describes the implemented database migration system, session state persistence, file management, and cleanup mechanisms for the OtherSide Paranormal Investigation Application.

## Overview

The application now includes a robust database management system with:

- **Database Migration System**: Automated schema versioning and migrations
- **Session State Persistence**: Persistent session state across application restarts
- **File Management System**: Complete file tracking with metadata and integrity checks
- **Cleanup Mechanisms**: Automated cleanup of old data and orphaned files

## Database Migration System

### Features

- **Version-based migrations** with tracking in `schema_migrations` table
- **Automatic migration discovery** from migration files
- **Rollback-safe** transaction-based migrations
- **Migration status tracking** and reporting

### Migration Files

Migrations are stored in `internal/repository/migrations/` with the naming format:
```
{version}_{description}.sql
```

Example:
- `001_initial_schema.sql` - Initial database schema
- `002_add_cleanup_indexes.sql` - Performance indexes for cleanup operations

### Commands

```bash
# Run migrations
./otherside -migrate

# Check migration status
./otherside -status

# Start application (auto-runs migrations)
./otherside
```

## Session State Persistence

### Features

- **Persistent session storage** in SQLite database
- **In-memory caching** of active sessions for performance
- **Session lifecycle management** (create, update, pause, resume, complete, archive)
- **Automatic cleanup** of expired sessions
- **Graceful shutdown** with state preservation

### Session States

- `active` - Currently running investigation
- `paused` - Temporarily paused investigation
- `complete` - Finished investigation with end time
- `archived` - Old sessions cleaned up but preserved

### Key Components

- `SessionStateManager` - Manages session state persistence
- `SQLiteSessionRepository` - Database operations for sessions
- Automatic recovery of active sessions on startup

## File Management System

### Features

- **File metadata tracking** with checksums and integrity verification
- **Support for multiple file types** (audio, video, data files)
- **Automatic MIME type detection**
- **SHA-256 checksums** for file integrity
- **Orphaned file detection** and cleanup
- **Storage quota enforcement**

### File Operations

- **Store**: Store files with automatic metadata generation
- **Retrieve**: Get files with metadata validation
- **Update**: Update existing files with new content
- **Delete**: Remove files and metadata
- **Verify**: Check file integrity against stored checksums

### Storage Structure

```
data/
├── sessions/
│   ├── {session-id}/
│   │   ├── recordings/
│   │   ├── exports/
│   │   └── media/
```

## Cleanup Mechanisms

### Features

- **Configurable retention policies** for sessions and files
- **Storage quota enforcement** with oldest file cleanup
- **Large file cleanup** based on size thresholds
- **Orphaned file cleanup** for files without metadata
- **Comprehensive reporting** of cleanup operations

### Cleanup Configuration

```go
type CleanupConfig struct {
    MaxSessionAge         time.Duration // Max age for old sessions
    MaxFileSizeBytes      int64         // Maximum individual file size
    MaxStorageSizeBytes   int64         // Total storage quota
    EnableFileCleanup     bool          // Enable file-based cleanup
    EnableDatabaseCleanup bool          // Enable database cleanup
    EnableOrphanCleanup   bool          // Enable orphaned file cleanup
}
```

### Commands

```bash
# Preview cleanup plan (doesn't delete anything)
./otherside -cleanup
# Then enter 'y' to proceed or 'n' to cancel
```

## Database Schema

### Tables

- `schema_migrations` - Migration tracking
- `sessions` - Investigation sessions
- `evp_recordings` - EVP recording data
- `vox_events` - VOX communication events
- `radar_events` - Radar detection events
- `sls_detections` - SLS detection data
- `user_interactions` - User interaction logs
- `files` - File metadata and tracking

### Indexes

Comprehensive indexing for performance:
- Session status and date indexes
- File size and creation date indexes
- Event timestamp indexes for time-series queries
- Foreign key indexes for join performance

## Configuration

### Database Configuration (Environment Variables)

```bash
DB_DRIVER=sqlite3           # Database driver
DB_NAME=otherside.db        # Database file name
DATA_PATH=./data            # Storage directory
MAX_SIZE_GB=10              # Maximum storage size
RETENTION_DAYS=30           # Session retention period
```

### Storage Configuration

```bash
AUDIO_SAMPLE_RATE=44100     # Audio processing config
MAX_RECORDING_MIN=30        # Maximum recording length
NOISE_THRESHOLD=0.1         # Noise detection threshold
```

## Initialization

### Application Startup

1. **Database Connection**: Initialize SQLite database
2. **Migration System**: Run pending migrations
3. **Session Recovery**: Load active sessions from database
4. **File Manager**: Initialize file storage and cleanup
5. **Cleanup Manager**: Configure cleanup policies
6. **Services**: Start application services

### Graceful Shutdown

1. **Session Persistence**: Save all active session states
2. **File Handles**: Close open file operations
3. **Database Connection**: Commit transactions and close
4. **Cleanup**: Final cleanup operations

## Usage Examples

### Session Management

```go
// Create new session
session := &domain.Session{
    ID:    "session-123",
    Title: "Paranormal Investigation at Location X",
    // ... other fields
}
err := sessionManager.CreateSession(ctx, session)

// Pause active session
err := sessionManager.PauseSession(ctx, "session-123")

// Resume paused session
err := sessionManager.ResumeSession(ctx, "session-123")

// Complete session
err := sessionManager.CompleteSession(ctx, "session-123")
```

### File Management

```go
// Store file
file, err := os.Open("recording.wav")
metadata, err := fileManager.StoreFile(ctx, sessionID, "recordings/audio.wav", file)

// Retrieve file
file, metadata, err := fileManager.GetFile(ctx, "recordings/audio.wav")

// Verify integrity
valid, err := fileManager.VerifyFileIntegrity(ctx, "recordings/audio.wav")
```

### Cleanup Operations

```go
// Configure cleanup
config := repository.CleanupConfig{
    MaxSessionAge:       30 * 24 * time.Hour, // 30 days
    MaxStorageSizeBytes: 10 * 1024 * 1024 * 1024, // 10GB
    EnableFileCleanup:   true,
    EnableDatabaseCleanup: true,
}

// Run cleanup
stats, err := cleanupManager.RunCleanup(ctx, config)
fmt.Printf("Cleaned %d sessions, %d files, freed %d bytes\n",
    stats.SessionsCleaned, stats.FilesCleaned, stats.BytesFreed)
```

## Error Handling

The system includes comprehensive error handling:

- **Transaction Rollback**: Failed migrations are rolled back automatically
- **Graceful Degradation**: Services continue operating with reduced functionality
- **Comprehensive Logging**: All operations are logged for debugging
- **Retry Logic**: Transient errors are handled with appropriate retries

## Security Considerations

- **File Isolation**: Files are stored in structured paths with session isolation
- **Checksum Verification**: All files have SHA-256 checksums for integrity
- **SQL Injection Protection**: All database queries use parameterized statements
- **Path Traversal Prevention**: File paths are validated and normalized

## Performance Optimization

- **Connection Pooling**: Database connections are properly managed
- **Indexing Strategy**: Comprehensive indexing for query performance
- **Batch Operations**: Cleanup operations use batch processing
- **Memory Caching**: Active sessions cached in memory for fast access

This system provides a robust foundation for the paranormal investigation application with data persistence, file management, and automated maintenance capabilities.