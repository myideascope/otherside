package domain

import (
	"context"
	"time"
)

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByID(ctx context.Context, id string) (*Session, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Session, error)
	GetByStatus(ctx context.Context, status SessionStatus) ([]*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id string) error
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*Session, error)
}

// EVPRepository defines the interface for EVP recording operations
type EVPRepository interface {
	Create(ctx context.Context, evp *EVPRecording) error
	GetByID(ctx context.Context, id string) (*EVPRecording, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*EVPRecording, error)
	Update(ctx context.Context, evp *EVPRecording) error
	Delete(ctx context.Context, id string) error
	GetByQuality(ctx context.Context, quality EVPQuality) ([]*EVPRecording, error)
	GetByDetectionLevel(ctx context.Context, minLevel float64) ([]*EVPRecording, error)
}

// VOXRepository defines the interface for VOX event operations
type VOXRepository interface {
	Create(ctx context.Context, vox *VOXEvent) error
	GetByID(ctx context.Context, id string) (*VOXEvent, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*VOXEvent, error)
	Update(ctx context.Context, vox *VOXEvent) error
	Delete(ctx context.Context, id string) error
	GetByLanguagePack(ctx context.Context, languagePack string) ([]*VOXEvent, error)
	GetByTriggerStrength(ctx context.Context, minStrength float64) ([]*VOXEvent, error)
}

// RadarRepository defines the interface for radar event operations
type RadarRepository interface {
	Create(ctx context.Context, radar *RadarEvent) error
	GetByID(ctx context.Context, id string) (*RadarEvent, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*RadarEvent, error)
	Update(ctx context.Context, radar *RadarEvent) error
	Delete(ctx context.Context, id string) error
	GetBySourceType(ctx context.Context, sourceType SourceType) ([]*RadarEvent, error)
	GetByStrengthRange(ctx context.Context, minStrength, maxStrength float64) ([]*RadarEvent, error)
}

// SLSRepository defines the interface for SLS detection operations
type SLSRepository interface {
	Create(ctx context.Context, sls *SLSDetection) error
	GetByID(ctx context.Context, id string) (*SLSDetection, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*SLSDetection, error)
	Update(ctx context.Context, sls *SLSDetection) error
	Delete(ctx context.Context, id string) error
	GetByConfidenceRange(ctx context.Context, minConfidence float64) ([]*SLSDetection, error)
	GetByDuration(ctx context.Context, minDuration float64) ([]*SLSDetection, error)
}

// InteractionRepository defines the interface for user interaction operations
type InteractionRepository interface {
	Create(ctx context.Context, interaction *UserInteraction) error
	GetByID(ctx context.Context, id string) (*UserInteraction, error)
	GetBySessionID(ctx context.Context, sessionID string) ([]*UserInteraction, error)
	Update(ctx context.Context, interaction *UserInteraction) error
	Delete(ctx context.Context, id string) error
	GetByType(ctx context.Context, interactionType InteractionType) ([]*UserInteraction, error)
}

// FileRepository defines the interface for file operations
type FileRepository interface {
	SaveFile(ctx context.Context, path string, data []byte) error
	GetFile(ctx context.Context, path string) ([]byte, error)
	DeleteFile(ctx context.Context, path string) error
	FileExists(ctx context.Context, path string) (bool, error)
	GetFileSize(ctx context.Context, path string) (int64, error)
	ListFiles(ctx context.Context, directory string) ([]string, error)
}
