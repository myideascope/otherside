-- Migration: 002_add_cleanup_indexes
-- Add indexes for better cleanup performance

CREATE INDEX IF NOT EXISTS idx_sessions_created_cleanup ON sessions(created_at, status);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
CREATE INDEX IF NOT EXISTS idx_files_file_size ON files(file_size);

-- Add file cleanup tracking
ALTER TABLE files ADD COLUMN last_accessed DATETIME DEFAULT CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_files_last_accessed ON files(last_accessed);