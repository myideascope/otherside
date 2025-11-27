package service

import (
	"context"
	"fmt"
	"time"

	"github.com/myideascope/otherside/internal/domain"
	"github.com/myideascope/otherside/pkg/audio"
)

// SessionService handles paranormal investigation session operations
type SessionService struct {
	sessionRepo     domain.SessionRepository
	evpRepo         domain.EVPRepository
	voxRepo         domain.VOXRepository
	radarRepo       domain.RadarRepository
	slsRepo         domain.SLSRepository
	interactionRepo domain.InteractionRepository
	fileRepo        domain.FileRepository
	audioProcessor  *audio.Processor
	voxGenerator    *audio.VOXGenerator
}

// SessionServiceConfig holds configuration for session service
type SessionServiceConfig struct {
	MaxConcurrentSessions int
	MaxSessionDuration    time.Duration
	StorageQuotaGB        int
}

// NewSessionService creates a new session service
func NewSessionService(
	sessionRepo domain.SessionRepository,
	evpRepo domain.EVPRepository,
	voxRepo domain.VOXRepository,
	radarRepo domain.RadarRepository,
	slsRepo domain.SLSRepository,
	interactionRepo domain.InteractionRepository,
	fileRepo domain.FileRepository,
	audioProcessor *audio.Processor,
	voxGenerator *audio.VOXGenerator,
) *SessionService {
	return &SessionService{
		sessionRepo:     sessionRepo,
		evpRepo:         evpRepo,
		voxRepo:         voxRepo,
		radarRepo:       radarRepo,
		slsRepo:         slsRepo,
		interactionRepo: interactionRepo,
		fileRepo:        fileRepo,
		audioProcessor:  audioProcessor,
		voxGenerator:    voxGenerator,
	}
}

// CreateSession creates a new paranormal investigation session
func (s *SessionService) CreateSession(ctx context.Context, req CreateSessionRequest) (*domain.Session, error) {
	session := &domain.Session{
		ID:            generateID(),
		Title:         req.Title,
		Location:      req.Location,
		StartTime:     time.Now(),
		Notes:         req.Notes,
		Environmental: req.Environmental,
		Status:        domain.SessionStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ProcessEVPRecording processes an EVP recording for paranormal analysis
func (s *SessionService) ProcessEVPRecording(ctx context.Context, sessionID string, audioData []float64, metadata EVPMetadata) (*domain.EVPRecording, error) {
	// Verify session exists and is active
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != domain.SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Process audio using audio processor
	result, err := s.audioProcessor.ProcessAudio(ctx, audioData)
	if err != nil {
		return nil, fmt.Errorf("audio processing failed: %w", err)
	}

	// Determine EVP quality based on anomaly strength and noise level
	quality := s.determineEVPQuality(result)

	// Create EVP recording
	evp := &domain.EVPRecording{
		ID:             generateID(),
		SessionID:      sessionID,
		FilePath:       metadata.FilePath,
		Duration:       result.Metadata.Duration,
		Timestamp:      time.Now(),
		WaveformData:   result.WaveformData,
		Annotations:    metadata.Annotations,
		Quality:        quality,
		DetectionLevel: result.AnomalyStrength,
		CreatedAt:      time.Now(),
	}

	if err := s.evpRepo.Create(ctx, evp); err != nil {
		return nil, fmt.Errorf("failed to save EVP recording: %w", err)
	}

	return evp, nil
}

// GenerateVOXCommunication generates VOX-based paranormal communication
func (s *SessionService) GenerateVOXCommunication(ctx context.Context, sessionID string, triggerData VOXTriggerData) (*domain.VOXEvent, error) {
	// Verify session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Ensure session is active
	if session.Status != domain.SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Prepare trigger data for VOX generator
	triggers := map[string]float64{
		"emf_anomaly":   triggerData.EMFAnomaly,
		"audio_anomaly": triggerData.AudioAnomaly,
		"temperature":   triggerData.TemperatureFluctuation,
		"interference":  triggerData.Interference,
	}

	// Generate VOX communication
	voxConfig := audio.VOXConfig{
		DefaultLanguage:  triggerData.LanguagePack,
		PhoneticBankSize: triggerData.PhoneticBankSize,
		TriggerThreshold: 0.3,
	}

	voxResult, err := s.voxGenerator.GenerateVOX(ctx, triggers, voxConfig)
	if err != nil {
		return nil, fmt.Errorf("VOX generation failed: %w", err)
	}

	// No generation if below threshold
	if voxResult == nil {
		return nil, nil
	}

	// Create VOX event
	voxEvent := &domain.VOXEvent{
		ID:              generateID(),
		SessionID:       sessionID,
		Timestamp:       time.Now(),
		GeneratedText:   voxResult.GeneratedText,
		PhoneticBank:    voxResult.PhoneticBank,
		FrequencyData:   voxResult.FrequencyData,
		TriggerStrength: voxResult.TriggerStrength,
		LanguagePack:    triggerData.LanguagePack,
		ModulationType:  voxResult.ModulationType,
		CreatedAt:       time.Now(),
	}

	if err := s.voxRepo.Create(ctx, voxEvent); err != nil {
		return nil, fmt.Errorf("failed to save VOX event: %w", err)
	}

	return voxEvent, nil
}

// ProcessRadarEvent processes radar/presence detection data
func (s *SessionService) ProcessRadarEvent(ctx context.Context, sessionID string, radarData RadarEventData) (*domain.RadarEvent, error) {
	// Verify session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Ensure session is active
	if session.Status != domain.SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Analyze radar data for authenticity (minimize false positives)
	if !s.validateRadarEvent(radarData) {
		return nil, fmt.Errorf("radar event failed validation")
	}

	// Determine source type based on data
	sourceType := s.determineRadarSourceType(radarData)

	radarEvent := &domain.RadarEvent{
		ID:            generateID(),
		SessionID:     sessionID,
		Timestamp:     time.Now(),
		Position:      radarData.Position,
		Strength:      radarData.Strength,
		SourceType:    sourceType,
		EMFReading:    radarData.EMFReading,
		AudioAnomaly:  radarData.AudioAnomaly,
		Duration:      radarData.Duration,
		MovementTrail: radarData.MovementTrail,
		CreatedAt:     time.Now(),
	}

	if err := s.radarRepo.Create(ctx, radarEvent); err != nil {
		return nil, fmt.Errorf("failed to save radar event: %w", err)
	}

	return radarEvent, nil
}

// ProcessSLSDetection processes SLS (Structured Light Sensor) detection
func (s *SessionService) ProcessSLSDetection(ctx context.Context, sessionID string, slsData SLSDetectionData) (*domain.SLSDetection, error) {
	// Verify session
	_, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Apply false-positive reduction filters
	if !s.validateSLSDetection(slsData) {
		return nil, fmt.Errorf("SLS detection failed validation")
	}

	// Analyze movement patterns
	movement := s.analyzeMovementPattern(slsData.SkeletalPoints)

	slsDetection := &domain.SLSDetection{
		ID:             generateID(),
		SessionID:      sessionID,
		Timestamp:      time.Now(),
		SkeletalPoints: slsData.SkeletalPoints,
		Confidence:     slsData.Confidence,
		BoundingBox:    slsData.BoundingBox,
		VideoFrame:     slsData.VideoFrame,
		FilterApplied:  slsData.FiltersApplied,
		Duration:       slsData.Duration,
		Movement:       movement,
		CreatedAt:      time.Now(),
	}

	if err := s.slsRepo.Create(ctx, slsDetection); err != nil {
		return nil, fmt.Errorf("failed to save SLS detection: %w", err)
	}

	return slsDetection, nil
}

// RecordUserInteraction records user interaction during investigation
func (s *SessionService) RecordUserInteraction(ctx context.Context, sessionID string, interaction UserInteractionData) (*domain.UserInteraction, error) {
	// Verify session
	_, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	userInteraction := &domain.UserInteraction{
		ID:               generateID(),
		SessionID:        sessionID,
		Timestamp:        time.Now(),
		Type:             interaction.Type,
		Content:          interaction.Content,
		AudioPath:        interaction.AudioPath,
		Response:         interaction.Response,
		ResponseTime:     interaction.ResponseTime,
		RandomizerResult: interaction.RandomizerResult,
		CreatedAt:        time.Now(),
	}

	if err := s.interactionRepo.Create(ctx, userInteraction); err != nil {
		return nil, fmt.Errorf("failed to save user interaction: %w", err)
	}

	return userInteraction, nil
}

// GetSessionSummary returns a comprehensive summary of a session
func (s *SessionService) GetSessionSummary(ctx context.Context, sessionID string) (*SessionSummary, error) {
	// Get session
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Get all related data
	evps, _ := s.evpRepo.GetBySessionID(ctx, sessionID)
	voxEvents, _ := s.voxRepo.GetBySessionID(ctx, sessionID)
	radarEvents, _ := s.radarRepo.GetBySessionID(ctx, sessionID)
	slsDetections, _ := s.slsRepo.GetBySessionID(ctx, sessionID)
	interactions, _ := s.interactionRepo.GetBySessionID(ctx, sessionID)

	// Calculate session statistics
	stats := s.calculateSessionStatistics(evps, voxEvents, radarEvents, slsDetections, interactions)

	return &SessionSummary{
		Session:       session,
		EVPs:          evps,
		VOXEvents:     voxEvents,
		RadarEvents:   radarEvents,
		SLSDetections: slsDetections,
		Interactions:  interactions,
		Statistics:    stats,
	}, nil
}

// Helper methods

func (s *SessionService) determineEVPQuality(result *audio.ProcessingResult) domain.EVPQuality {
	if result.AnomalyStrength >= 0.8 && result.NoiseLevel < 0.1 {
		return domain.EVPQualityExcellent
	} else if result.AnomalyStrength >= 0.6 && result.NoiseLevel < 0.2 {
		return domain.EVPQualityGood
	} else if result.AnomalyStrength >= 0.4 && result.NoiseLevel < 0.4 {
		return domain.EVPQualityFair
	}
	return domain.EVPQualityPoor
}

func (s *SessionService) validateRadarEvent(data RadarEventData) bool {
	// Implement validation logic to minimize false positives
	// Check for minimum strength threshold
	if data.Strength < 0.3 {
		return false
	}

	// Validate position data
	if data.Position.X == 0 && data.Position.Y == 0 {
		return false
	}

	// Check for reasonable EMF readings
	if data.EMFReading < 0 || data.EMFReading > 1000 {
		return false
	}

	return true
}

func (s *SessionService) determineRadarSourceType(data RadarEventData) domain.SourceType {
	if data.EMFReading > 0.5 && data.AudioAnomaly > 0.5 {
		return domain.SourceTypeBoth
	} else if data.EMFReading > 0.5 {
		return domain.SourceTypeEMF
	} else if data.AudioAnomaly > 0.5 {
		return domain.SourceTypeAudio
	}
	return domain.SourceTypeOther
}

func (s *SessionService) validateSLSDetection(data SLSDetectionData) bool {
	// Minimum confidence threshold
	if data.Confidence < 0.5 {
		return false
	}

	// Minimum number of skeletal points
	if len(data.SkeletalPoints) < 5 {
		return false
	}

	// Validate bounding box
	if data.BoundingBox.Width < 10 || data.BoundingBox.Height < 10 {
		return false
	}

	return true
}

func (s *SessionService) analyzeMovementPattern(points []domain.SkeletalPoint) domain.MovementAnalysis {
	if len(points) < 2 {
		return domain.MovementAnalysis{
			Speed:     0,
			Direction: 0,
			Pattern:   "static",
		}
	}

	// Simple movement analysis
	// In a real implementation, this would be more sophisticated
	return domain.MovementAnalysis{
		Speed:     0.5, // Placeholder
		Direction: 0,   // Placeholder
		Pattern:   "unknown",
	}
}

func (s *SessionService) calculateSessionStatistics(
	evps []*domain.EVPRecording,
	voxEvents []*domain.VOXEvent,
	radarEvents []*domain.RadarEvent,
	slsDetections []*domain.SLSDetection,
	interactions []*domain.UserInteraction,
) SessionStatistics {
	stats := SessionStatistics{
		TotalEVPs:          len(evps),
		TotalVOXEvents:     len(voxEvents),
		TotalRadarEvents:   len(radarEvents),
		TotalSLSDetections: len(slsDetections),
		TotalInteractions:  len(interactions),
	}

	// Calculate EVP quality distribution
	for _, evp := range evps {
		switch evp.Quality {
		case domain.EVPQualityExcellent:
			stats.HighQualityEVPs++
		case domain.EVPQualityGood:
			stats.MediumQualityEVPs++
		}
	}

	// Calculate average anomaly strength
	if len(evps) > 0 {
		var totalStrength float64
		for _, evp := range evps {
			totalStrength += evp.DetectionLevel
		}
		stats.AverageAnomalyStrength = totalStrength / float64(len(evps))
	}

	return stats
}

// generateID generates a unique ID (simplified implementation)
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Request/Response types

type CreateSessionRequest struct {
	Title         string               `json:"title"`
	Location      domain.Location      `json:"location"`
	Notes         string               `json:"notes"`
	Environmental domain.Environmental `json:"environmental"`
}

type EVPMetadata struct {
	FilePath    string   `json:"file_path"`
	Annotations []string `json:"annotations"`
}

type VOXTriggerData struct {
	EMFAnomaly             float64 `json:"emf_anomaly"`
	AudioAnomaly           float64 `json:"audio_anomaly"`
	TemperatureFluctuation float64 `json:"temperature_fluctuation"`
	Interference           float64 `json:"interference"`
	LanguagePack           string  `json:"language_pack"`
	PhoneticBankSize       int     `json:"phonetic_bank_size"`
}

type RadarEventData struct {
	Position      domain.Coordinates   `json:"position"`
	Strength      float64              `json:"strength"`
	EMFReading    float64              `json:"emf_reading"`
	AudioAnomaly  float64              `json:"audio_anomaly"`
	Duration      float64              `json:"duration"`
	MovementTrail []domain.Coordinates `json:"movement_trail"`
}

type SLSDetectionData struct {
	SkeletalPoints []domain.SkeletalPoint `json:"skeletal_points"`
	Confidence     float64                `json:"confidence"`
	BoundingBox    domain.BoundingBox     `json:"bounding_box"`
	VideoFrame     string                 `json:"video_frame"`
	FiltersApplied []string               `json:"filters_applied"`
	Duration       float64                `json:"duration"`
}

type UserInteractionData struct {
	Type             domain.InteractionType   `json:"type"`
	Content          string                   `json:"content"`
	AudioPath        string                   `json:"audio_path,omitempty"`
	Response         string                   `json:"response,omitempty"`
	ResponseTime     float64                  `json:"response_time,omitempty"`
	RandomizerResult *domain.RandomizerResult `json:"randomizer_result,omitempty"`
}

type SessionSummary struct {
	Session       *domain.Session           `json:"session"`
	EVPs          []*domain.EVPRecording    `json:"evps"`
	VOXEvents     []*domain.VOXEvent        `json:"vox_events"`
	RadarEvents   []*domain.RadarEvent      `json:"radar_events"`
	SLSDetections []*domain.SLSDetection    `json:"sls_detections"`
	Interactions  []*domain.UserInteraction `json:"interactions"`
	Statistics    SessionStatistics         `json:"statistics"`
}

type SessionStatistics struct {
	TotalEVPs              int     `json:"total_evps"`
	TotalVOXEvents         int     `json:"total_vox_events"`
	TotalRadarEvents       int     `json:"total_radar_events"`
	TotalSLSDetections     int     `json:"total_sls_detections"`
	TotalInteractions      int     `json:"total_interactions"`
	HighQualityEVPs        int     `json:"high_quality_evps"`
	MediumQualityEVPs      int     `json:"medium_quality_evps"`
	AverageAnomalyStrength float64 `json:"average_anomaly_strength"`
}
