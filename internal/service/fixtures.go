package service

import (
	"time"

	"github.com/myideascope/otherside/internal/domain"
	"github.com/myideascope/otherside/pkg/audio"
)

// TestSession provides a complete test session
func TestSession() *domain.Session {
	return &domain.Session{
		ID:            "test-session-123",
		Title:         "Test Investigation",
		Location:      TestLocation(),
		StartTime:     time.Now().Add(-2 * time.Hour),
		Notes:         "Test session for unit testing",
		Environmental: TestEnvironmental(),
		Status:        domain.SessionStatusActive,
		CreatedAt:     time.Now().Add(-2 * time.Hour),
		UpdatedAt:     time.Now().Add(-2 * time.Hour),
	}
}

// TestLocation provides a test location
func TestLocation() domain.Location {
	return domain.Location{
		Latitude:    37.7749,
		Longitude:   -122.4194,
		Address:     "123 Test Street, San Francisco, CA",
		Description: "Test haunted location",
		Venue:       "Test Mansion",
	}
}

// TestEnvironmental provides test environmental data
func TestEnvironmental() domain.Environmental {
	return domain.Environmental{
		Temperature: 20.5,
		Humidity:    45.0,
		Pressure:    1013.25,
		EMFLevel:    2.5,
		LightLevel:  150.0,
		NoiseLevel:  0.3,
	}
}

// TestEVPRecording provides a test EVP recording
func TestEVPRecording() *domain.EVPRecording {
	return &domain.EVPRecording{
		ID:             "test-evp-456",
		SessionID:      "test-session-123",
		FilePath:       "/audio/test_evp.wav",
		Duration:       3.5,
		Timestamp:      time.Now().Add(-1 * time.Hour),
		WaveformData:   []float64{0.1, 0.2, 0.3, 0.2, 0.1, -0.1, -0.2, -0.3},
		ProcessedPath:  "/audio/test_evp_processed.wav",
		Annotations:    []string{"Possible voice", "Low frequency anomaly"},
		Quality:        domain.EVPQualityGood,
		DetectionLevel: 0.75,
		CreatedAt:      time.Now().Add(-1 * time.Hour),
	}
}

// TestVOXEvent provides a test VOX event
func TestVOXEvent() *domain.VOXEvent {
	return &domain.VOXEvent{
		ID:              "test-vox-789",
		SessionID:       "test-session-123",
		Timestamp:       time.Now().Add(-30 * time.Minute),
		GeneratedText:   "hello there",
		PhoneticBank:    "english",
		FrequencyData:   []float64{440.0, 880.0, 1320.0, 1760.0, 2200.0},
		TriggerStrength: 0.85,
		LanguagePack:    "english",
		ModulationType:  "amplitude",
		UserResponse:    "I can hear you!",
		ResponseDelay:   2.5,
		CreatedAt:       time.Now().Add(-30 * time.Minute),
	}
}

// TestRadarEvent provides a test radar event
func TestRadarEvent() *domain.RadarEvent {
	return &domain.RadarEvent{
		ID:           "test-radar-012",
		SessionID:    "test-session-123",
		Timestamp:    time.Now().Add(-45 * time.Minute),
		Position:     domain.Coordinates{X: 5.5, Y: 3.2, Z: 1.8},
		Strength:     0.78,
		SourceType:   domain.SourceTypeBoth,
		EMFReading:   4.2,
		AudioAnomaly: 0.65,
		Duration:     12.5,
		MovementTrail: []domain.Coordinates{
			{X: 5.0, Y: 3.0, Z: 1.5},
			{X: 5.5, Y: 3.2, Z: 1.8},
			{X: 6.0, Y: 3.4, Z: 2.0},
		},
		CreatedAt: time.Now().Add(-45 * time.Minute),
	}
}

// TestSLSDetection provides a test SLS detection
func TestSLSDetection() *domain.SLSDetection {
	return &domain.SLSDetection{
		ID:        "test-sls-345",
		SessionID: "test-session-123",
		Timestamp: time.Now().Add(-20 * time.Minute),
		SkeletalPoints: []domain.SkeletalPoint{
			{Joint: "head", Position: domain.Coordinates{X: 0.0, Y: 1.8, Z: 0.0}, Confidence: 0.92},
			{Joint: "shoulder_left", Position: domain.Coordinates{X: -0.3, Y: 1.4, Z: 0.1}, Confidence: 0.88},
			{Joint: "shoulder_right", Position: domain.Coordinates{X: 0.3, Y: 1.4, Z: 0.1}, Confidence: 0.89},
			{Joint: "hip_left", Position: domain.Coordinates{X: -0.2, Y: 0.8, Z: 0.0}, Confidence: 0.85},
			{Joint: "hip_right", Position: domain.Coordinates{X: 0.2, Y: 0.8, Z: 0.0}, Confidence: 0.86},
			{Joint: "knee_left", Position: domain.Coordinates{X: -0.2, Y: 0.4, Z: 0.0}, Confidence: 0.83},
			{Joint: "knee_right", Position: domain.Coordinates{X: 0.2, Y: 0.4, Z: 0.0}, Confidence: 0.84},
		},
		Confidence: 0.87,
		BoundingBox: domain.BoundingBox{
			TopLeft:     domain.Coordinates{X: -0.5, Y: 0.0, Z: -0.2},
			BottomRight: domain.Coordinates{X: 0.5, Y: 2.0, Z: 0.2},
			Width:       1.0,
			Height:      2.0,
		},
		VideoFrame:    "/video/frame_001.jpg",
		FilterApplied: []string{"noise_reduction", "motion_stabilization"},
		Duration:      5.2,
		Movement: domain.MovementAnalysis{
			Speed:     0.15,
			Direction: 45.0,
			Pattern:   "slow_drift",
		},
		CreatedAt: time.Now().Add(-20 * time.Minute),
	}
}

// TestUserInteraction provides a test user interaction
func TestUserInteraction() *domain.UserInteraction {
	return &domain.UserInteraction{
		ID:           "test-interaction-678",
		SessionID:    "test-session-123",
		Timestamp:    time.Now().Add(-10 * time.Minute),
		Type:         domain.InteractionTypeVoice,
		Content:      "Is anyone there? I think I heard something.",
		AudioPath:    "/audio/user_interaction_001.wav",
		Response:     "The temperature just dropped suddenly",
		ResponseTime: 8.5,
		RandomizerResult: &domain.RandomizerResult{
			Type:   "dice",
			Result: int64(6),
			Range:  "1-6",
		},
		CreatedAt: time.Now().Add(-10 * time.Minute),
	}
}

// TestAudioData provides sample audio data for testing
func TestAudioData() []float64 {
	return []float64{
		0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
		0.9, 0.8, 0.7, 0.6, 0.5, 0.4, 0.3, 0.2, 0.1, 0.0,
		-0.1, -0.2, -0.3, -0.4, -0.5, -0.6, -0.7, -0.8, -0.9, -1.0,
		-0.9, -0.8, -0.7, -0.6, -0.5, -0.4, -0.3, -0.2, -0.1, 0.0,
		0.1, 0.3, 0.5, 0.7, 0.9, 0.7, 0.5, 0.3, 0.1, -0.1,
	}
}

// TestProcessingResult provides a test audio processing result
func TestProcessingResult() *audio.ProcessingResult {
	return &audio.ProcessingResult{
		WaveformData:  TestAudioData(),
		FrequencyData: []complex128{complex(1.0, 0.0), complex(0.5, 0.5), complex(0.0, 1.0), complex(-0.5, 0.5)},
		EVPEvents: []audio.EVPEvent{
			{
				StartTime:   1.0,
				EndTime:     1.5,
				Confidence:  0.8,
				Frequency:   440.0,
				Amplitude:   0.6,
				Description: "Voice-like anomaly detected",
			},
		},
		AnomalyStrength: 0.75,
		NoiseLevel:      0.12,
		ProcessingTime:  150 * time.Millisecond,
		SpectralAnalysis: audio.SpectralAnalysis{
			DominantFrequencies: []audio.FrequencyPeak{
				{Frequency: 440.0, Magnitude: 0.8, Quality: 0.9},
				{Frequency: 880.0, Magnitude: 0.4, Quality: 0.7},
			},
			SpectralCentroid: 650.0,
			SpectralRolloff:  1200.0,
			ZeroCrossingRate: 0.15,
			RMSEnergy:        0.45,
		},
		Metadata: audio.ProcessingMetadata{
			SampleRate:  44100,
			BitDepth:    16,
			Duration:    1.5,
			ProcessedAt: time.Now(),
			FilterSettings: audio.FilterSettings{
				HighPassCutoff: 80.0,
				LowPassCutoff:  8000.0,
				NotchFilters:   []float64{50.0, 60.0},
				NoiseReduction: true,
				DynamicRange:   false,
			},
		},
	}
}

// TestVOXResult provides a test VOX generation result
func TestVOXResult() *audio.VOXResult {
	return &audio.VOXResult{
		GeneratedText:   "hello there",
		PhoneticBank:    "english",
		TriggerStrength: 0.85,
		FrequencyData:   []float64{440.0, 880.0, 1320.0, 1760.0, 2200.0},
		ModulationType:  "amplitude",
		GeneratedAt:     time.Now(),
	}
}

// TestVOXConfig provides a test VOX configuration
func TestVOXConfig() audio.VOXConfig {
	return audio.VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 23,
		TriggerThreshold: 0.3,
	}
}

// TestCreateSessionRequest provides a test session creation request
func TestCreateSessionRequest() CreateSessionRequest {
	return CreateSessionRequest{
		Title:         "Test Investigation",
		Location:      TestLocation(),
		Notes:         "Test session for unit testing",
		Environmental: TestEnvironmental(),
	}
}

// TestEVPMetadata provides test EVP metadata
func TestEVPMetadata() EVPMetadata {
	return EVPMetadata{
		FilePath:    "/audio/test_evp.wav",
		Annotations: []string{"Possible voice", "Low frequency anomaly"},
	}
}

// TestVOXTriggerData provides test VOX trigger data
func TestVOXTriggerData() VOXTriggerData {
	return VOXTriggerData{
		EMFAnomaly:             0.75,
		AudioAnomaly:           0.65,
		TemperatureFluctuation: 0.15,
		Interference:           0.25,
		LanguagePack:           "english",
		PhoneticBankSize:       23,
	}
}

// TestRadarEventData provides test radar event data
func TestRadarEventData() RadarEventData {
	return RadarEventData{
		Position:     domain.Coordinates{X: 5.5, Y: 3.2, Z: 1.8},
		Strength:     0.78,
		EMFReading:   4.2,
		AudioAnomaly: 0.65,
		Duration:     12.5,
		MovementTrail: []domain.Coordinates{
			{X: 5.0, Y: 3.0, Z: 1.5},
			{X: 5.5, Y: 3.2, Z: 1.8},
			{X: 6.0, Y: 3.4, Z: 2.0},
		},
	}
}

// TestSLSDetectionData provides test SLS detection data
func TestSLSDetectionData() SLSDetectionData {
	return SLSDetectionData{
		SkeletalPoints: []domain.SkeletalPoint{
			{Joint: "head", Position: domain.Coordinates{X: 0.0, Y: 1.8, Z: 0.0}, Confidence: 0.92},
			{Joint: "shoulder_left", Position: domain.Coordinates{X: -0.3, Y: 1.4, Z: 0.1}, Confidence: 0.88},
			{Joint: "shoulder_right", Position: domain.Coordinates{X: 0.3, Y: 1.4, Z: 0.1}, Confidence: 0.89},
			{Joint: "hip_left", Position: domain.Coordinates{X: -0.2, Y: 0.8, Z: 0.0}, Confidence: 0.85},
			{Joint: "hip_right", Position: domain.Coordinates{X: 0.2, Y: 0.8, Z: 0.0}, Confidence: 0.86},
			{Joint: "knee_left", Position: domain.Coordinates{X: -0.2, Y: 0.4, Z: 0.0}, Confidence: 0.83},
			{Joint: "knee_right", Position: domain.Coordinates{X: 0.2, Y: 0.4, Z: 0.0}, Confidence: 0.84},
		},
		Confidence: 0.87,
		BoundingBox: domain.BoundingBox{
			TopLeft:     domain.Coordinates{X: -0.5, Y: 0.0, Z: -0.2},
			BottomRight: domain.Coordinates{X: 0.5, Y: 2.0, Z: 0.2},
			Width:       1.0,
			Height:      2.0,
		},
		VideoFrame:     "/video/frame_001.jpg",
		FiltersApplied: []string{"noise_reduction", "motion_stabilization"},
		Duration:       5.2,
	}
}

// TestUserInteractionData provides test user interaction data
func TestUserInteractionData() UserInteractionData {
	return UserInteractionData{
		Type:         domain.InteractionTypeVoice,
		Content:      "Is anyone there? I think I heard something.",
		AudioPath:    "/audio/user_interaction_001.wav",
		Response:     "The temperature just dropped suddenly",
		ResponseTime: 8.5,
		RandomizerResult: &domain.RandomizerResult{
			Type:   "dice",
			Result: int64(6),
			Range:  "1-6",
		},
	}
}

// TestExportRequest provides a test export request
func TestExportRequest() ExportRequest {
	return ExportRequest{
		SessionIDs:   []string{"test-session-123", "test-session-456"},
		Format:       ExportFormatJSON,
		IncludeAudio: true,
		IncludeVideo: false,
		DateFrom:     nil,
		DateTo:       nil,
		IncludeEVPs:  true,
		IncludeVOX:   true,
		IncludeRadar: true,
		IncludeSLS:   true,
		IncludeNotes: true,
	}
}
