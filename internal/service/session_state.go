package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/myideascope/otherside/internal/domain"
	"github.com/myideascope/otherside/internal/repository"
)

// SessionStateManager manages session state persistence across restarts
type SessionStateManager struct {
	sessionRepo    *repository.SQLiteSessionRepository
	activeSessions map[string]*domain.Session
	mu             sync.RWMutex
	db             *sql.DB
}

// NewSessionStateManager creates a new session state manager
func NewSessionStateManager(db *sql.DB) *SessionStateManager {
	return &SessionStateManager{
		sessionRepo:    repository.NewSQLiteSessionRepository(db),
		activeSessions: make(map[string]*domain.Session),
		db:             db,
	}
}

// Initialize loads active sessions from database
func (sm *SessionStateManager) Initialize(ctx context.Context) error {
	log.Println("Loading active sessions from database...")

	// Load all active sessions
	sessions, err := sm.sessionRepo.GetByStatus(ctx, domain.SessionStatusActive)
	if err != nil {
		return fmt.Errorf("failed to load active sessions: %w", err)
	}

	// Load paused sessions as well
	pausedSessions, err := sm.sessionRepo.GetByStatus(ctx, domain.SessionStatusPaused)
	if err != nil {
		return fmt.Errorf("failed to load paused sessions: %w", err)
	}

	sessions = append(sessions, pausedSessions...)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, session := range sessions {
		sm.activeSessions[session.ID] = session
		log.Printf("Loaded session: %s (%s)", session.ID, session.Status)
	}

	log.Printf("Loaded %d sessions from database", len(sm.activeSessions))
	return nil
}

// CreateSession creates and persists a new session
func (sm *SessionStateManager) CreateSession(ctx context.Context, session *domain.Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Set timestamps
	now := time.Now()
	session.CreatedAt = now
	session.UpdatedAt = now
	session.Status = domain.SessionStatusActive

	// Persist to database
	if err := sm.sessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session in database: %w", err)
	}

	// Add to active sessions
	sm.activeSessions[session.ID] = session
	log.Printf("Created new session: %s", session.ID)

	return nil
}

// GetSession retrieves a session by ID (first from memory, then from database)
func (sm *SessionStateManager) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	// Try memory first
	sm.mu.RLock()
	if session, exists := sm.activeSessions[id]; exists {
		sm.mu.RUnlock()
		return session, nil
	}
	sm.mu.RUnlock()

	// Load from database
	session, err := sm.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from database: %w", err)
	}

	// Cache if active or paused
	if session.Status == domain.SessionStatusActive || session.Status == domain.SessionStatusPaused {
		sm.mu.Lock()
		sm.activeSessions[id] = session
		sm.mu.Unlock()
	}

	return session, nil
}

// UpdateSession updates and persists a session
func (sm *SessionStateManager) UpdateSession(ctx context.Context, session *domain.Session) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Update timestamp
	session.UpdatedAt = time.Now()

	// Persist to database
	if err := sm.sessionRepo.Update(ctx, session); err != nil {
		return fmt.Errorf("failed to update session in database: %w", err)
	}

	// Update in memory
	if _, exists := sm.activeSessions[session.ID]; exists {
		sm.activeSessions[session.ID] = session
	}

	// If session is completed or archived, remove from active sessions
	if session.Status == domain.SessionStatusComplete || session.Status == domain.SessionStatusArchived {
		delete(sm.activeSessions, session.ID)
	}

	log.Printf("Updated session: %s (status: %s)", session.ID, session.Status)
	return nil
}

// PauseSession pauses a session
func (sm *SessionStateManager) PauseSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Status = domain.SessionStatusPaused
	return sm.UpdateSession(ctx, session)
}

// ResumeSession resumes a paused session
func (sm *SessionStateManager) ResumeSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Status = domain.SessionStatusActive
	return sm.UpdateSession(ctx, session)
}

// CompleteSession marks a session as complete
func (sm *SessionStateManager) CompleteSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.EndTime = &now
	session.Status = domain.SessionStatusComplete
	return sm.UpdateSession(ctx, session)
}

// ArchiveSession archives a session
func (sm *SessionStateManager) ArchiveSession(ctx context.Context, sessionID string) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Status = domain.SessionStatusArchived
	return sm.UpdateSession(ctx, session)
}

// DeleteSession deletes a session completely
func (sm *SessionStateManager) DeleteSession(ctx context.Context, sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Delete from database
	if err := sm.sessionRepo.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session from database: %w", err)
	}

	// Remove from memory
	delete(sm.activeSessions, sessionID)
	log.Printf("Deleted session: %s", sessionID)

	return nil
}

// GetActiveSessions returns all active sessions
func (sm *SessionStateManager) GetActiveSessions() []*domain.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*domain.Session, 0, len(sm.activeSessions))
	for _, session := range sm.activeSessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// GetActiveSessionsByStatus returns sessions by status
func (sm *SessionStateManager) GetActiveSessionsByStatus(status domain.SessionStatus) []*domain.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*domain.Session
	for _, session := range sm.activeSessions {
		if session.Status == status {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// SaveSessionState saves the current state of all active sessions
func (sm *SessionStateManager) SaveSessionState(ctx context.Context) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	log.Printf("Saving state for %d active sessions", len(sm.activeSessions))

	for _, session := range sm.activeSessions {
		if err := sm.sessionRepo.Update(ctx, session); err != nil {
			log.Printf("Failed to save session %s: %v", session.ID, err)
			return fmt.Errorf("failed to save session %s: %w", session.ID, err)
		}
	}

	log.Println("All session states saved successfully")
	return nil
}

// CleanupExpiredSessions cleans up sessions that have been inactive for too long
func (sm *SessionStateManager) CleanupExpiredSessions(ctx context.Context, maxInactiveDuration time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var expiredSessions []string

	for id, session := range sm.activeSessions {
		inactiveDuration := now.Sub(session.UpdatedAt)

		// Consider sessions expired if inactive for too long and not already completed
		if inactiveDuration > maxInactiveDuration &&
			(session.Status == domain.SessionStatusActive || session.Status == domain.SessionStatusPaused) {
			expiredSessions = append(expiredSessions, id)
		}
	}

	// Mark expired sessions as archived
	for _, id := range expiredSessions {
		session := sm.activeSessions[id]
		session.Status = domain.SessionStatusArchived
		session.UpdatedAt = now

		if err := sm.sessionRepo.Update(ctx, session); err != nil {
			log.Printf("Failed to archive expired session %s: %v", id, err)
			continue
		}

		delete(sm.activeSessions, id)
		log.Printf("Archived expired session: %s (inactive for %v)", id, maxInactiveDuration)
	}

	if len(expiredSessions) > 0 {
		log.Printf("Cleaned up %d expired sessions", len(expiredSessions))
	}

	return nil
}

// Shutdown gracefully shuts down the session manager
func (sm *SessionStateManager) Shutdown(ctx context.Context) error {
	log.Println("Shutting down session state manager...")

	// Save all active session states
	if err := sm.SaveSessionState(ctx); err != nil {
		return fmt.Errorf("failed to save session states during shutdown: %w", err)
	}

	log.Println("Session state manager shutdown complete")
	return nil
}
