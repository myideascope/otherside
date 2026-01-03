package domain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEVPQuality_Values(t *testing.T) {
	tests := []struct {
		name     string
		quality  EVPQuality
		expected string
	}{
		{"Excellent", EVPQualityExcellent, "excellent"},
		{"Good", EVPQualityGood, "good"},
		{"Fair", EVPQualityFair, "fair"},
		{"Poor", EVPQualityPoor, "poor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.quality))
		})
	}
}

func TestEVPRecording_Validation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		evp       EVPRecording
		expectErr bool
	}{
		{
			name: "ValidEVPRecording",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3, 0.4, 0.5},
				Annotations:    []string{"Voice detected", "Low frequency"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: false,
		},
		{
			name: "MissingID",
			evp: EVPRecording{
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "MissingSessionID",
			evp: EVPRecording{
				ID:             "evp-123",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "EmptyFilePath",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "NegativeDuration",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       -5.0,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "EmptyWaveformData",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "NilWaveformData",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   nil,
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "DetectionLevelTooLow",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: -0.1,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "DetectionLevelTooHigh",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityGood,
				DetectionLevel: 1.1,
				CreatedAt:      now,
			},
			expectErr: true,
		},
		{
			name: "BoundaryDetectionLevelZero",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityPoor,
				DetectionLevel: 0.0,
				CreatedAt:      now,
			},
			expectErr: false,
		},
		{
			name: "BoundaryDetectionLevelOne",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Voice detected"},
				Quality:        EVPQualityExcellent,
				DetectionLevel: 1.0,
				CreatedAt:      now,
			},
			expectErr: false,
		},
		{
			name: "ValidWithEmptyAnnotations",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: false,
		},
		{
			name: "ValidWithNilAnnotations",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    nil,
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: false,
		},
		{
			name: "ValidWithProcessedPath",
			evp: EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				ProcessedPath:  "/path/to/processed.wav",
				Annotations:    []string{"Processed"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.evp.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEVPRecording_JSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	evp := EVPRecording{
		ID:             "evp-123",
		SessionID:      "session-456",
		FilePath:       "/path/to/recording.wav",
		Duration:       30.5,
		Timestamp:      now,
		WaveformData:   []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		ProcessedPath:  "/path/to/processed.wav",
		Annotations:    []string{"Voice detected", "Low frequency", "Echo"},
		Quality:        EVPQualityExcellent,
		DetectionLevel: 0.85,
		CreatedAt:      now,
	}

	t.Run("MarshalJSON", func(t *testing.T) {
		data, err := json.Marshal(evp)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		var unmarshaled EVPRecording
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, evp.ID, unmarshaled.ID)
		assert.Equal(t, evp.SessionID, unmarshaled.SessionID)
		assert.Equal(t, evp.FilePath, unmarshaled.FilePath)
		assert.Equal(t, evp.Duration, unmarshaled.Duration)
		assert.Equal(t, evp.Quality, unmarshaled.Quality)
		assert.Equal(t, evp.DetectionLevel, unmarshaled.DetectionLevel)
		assert.Equal(t, evp.Timestamp.Unix(), unmarshaled.Timestamp.Unix())
		assert.Equal(t, evp.CreatedAt.Unix(), unmarshaled.CreatedAt.Unix())
		assert.Equal(t, evp.ProcessedPath, unmarshaled.ProcessedPath)
		assert.Equal(t, evp.WaveformData, unmarshaled.WaveformData)
		assert.Equal(t, evp.Annotations, unmarshaled.Annotations)
	})

	t.Run("MarshalJSONWithoutProcessedPath", func(t *testing.T) {
		evpWithoutProcessed := EVPRecording{
			ID:             "evp-123",
			SessionID:      "session-456",
			FilePath:       "/path/to/recording.wav",
			Duration:       30.5,
			Timestamp:      now,
			WaveformData:   []float64{0.1, 0.2, 0.3},
			Annotations:    []string{"Basic recording"},
			Quality:        EVPQualityFair,
			DetectionLevel: 0.6,
			CreatedAt:      now,
		}

		data, err := json.Marshal(evpWithoutProcessed)
		require.NoError(t, err)

		var unmarshaled EVPRecording
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, evpWithoutProcessed.ID, unmarshaled.ID)
		assert.Equal(t, evpWithoutProcessed.FilePath, unmarshaled.FilePath)
		assert.Empty(t, unmarshaled.ProcessedPath, "ProcessedPath should be empty when not set")
	})

	t.Run("MarshalJSONWithEmptyAnnotations", func(t *testing.T) {
		evpWithEmptyAnnotations := EVPRecording{
			ID:             "evp-123",
			SessionID:      "session-456",
			FilePath:       "/path/to/recording.wav",
			Duration:       30.5,
			Timestamp:      now,
			WaveformData:   []float64{0.1, 0.2, 0.3},
			Annotations:    []string{},
			Quality:        EVPQualityPoor,
			DetectionLevel: 0.25,
			CreatedAt:      now,
		}

		data, err := json.Marshal(evpWithEmptyAnnotations)
		require.NoError(t, err)

		var unmarshaled EVPRecording
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, evpWithEmptyAnnotations.ID, unmarshaled.ID)
		assert.Equal(t, evpWithEmptyAnnotations.Annotations, unmarshaled.Annotations)
	})
}

func TestEVPRecording_DatabaseTags(t *testing.T) {
	evp := EVPRecording{}
	evpType := reflect.TypeOf(evp)

	requiredTags := []string{"id", "session_id", "file_path", "duration", "timestamp", "waveform_data", "quality", "detection_level", "created_at"}
	optionalTags := []string{"processed_path", "annotations"}

	for _, tag := range requiredTags {
		field, found := evpType.FieldByNameFunc(func(name string) bool {
			field, _ := evpType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		assert.True(t, found, "Required field with db tag '%s' not found", tag)
		assert.NotEmpty(t, field.Name, "Field name should not be empty for tag '%s'", tag)
	}

	for _, tag := range optionalTags {
		field, found := evpType.FieldByNameFunc(func(name string) bool {
			field, _ := evpType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		if found {
			assert.NotEmpty(t, field.Name, "Field name should not be empty for optional tag '%s'", tag)
		}
	}
}

func TestEVPRecording_AnnotationHandling(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		annotations  []string
		expectValid  bool
		expectedSize int
	}{
		{
			name:         "MultipleAnnotations",
			annotations:  []string{"Voice detected", "Low frequency", "Echo pattern", "Static noise"},
			expectValid:  true,
			expectedSize: 4,
		},
		{
			name:         "SingleAnnotation",
			annotations:  []string{"Single voice detection"},
			expectValid:  true,
			expectedSize: 1,
		},
		{
			name:         "EmptyAnnotations",
			annotations:  []string{},
			expectValid:  true,
			expectedSize: 0,
		},
		{
			name:         "NilAnnotations",
			annotations:  nil,
			expectValid:  true,
			expectedSize: 0,
		},
		{
			name:         "AnnotationWithEmptyString",
			annotations:  []string{"Voice detected", ""},
			expectValid:  true,
			expectedSize: 2,
		},
		{
			name:         "LongAnnotation",
			annotations:  []string{"This is a very long annotation that describes in great detail the paranormal phenomena detected in the EVP recording, including specific time stamps and frequency analysis results"},
			expectValid:  true,
			expectedSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evp := EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    tt.annotations,
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			}

			err := evp.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSize, len(evp.Annotations))
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEVPRecording_WaveformValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		waveformData   []float64
		expectValid    bool
		expectedSize   int
		expectedValues []float64
	}{
		{
			name:           "ValidWaveform",
			waveformData:   []float64{0.1, -0.2, 0.3, -0.4, 0.5},
			expectValid:    true,
			expectedSize:   5,
			expectedValues: []float64{0.1, -0.2, 0.3, -0.4, 0.5},
		},
		{
			name:           "SingleSampleWaveform",
			waveformData:   []float64{0.0},
			expectValid:    true,
			expectedSize:   1,
			expectedValues: []float64{0.0},
		},
		{
			name:           "EmptyWaveform",
			waveformData:   []float64{},
			expectValid:    false,
			expectedSize:   0,
			expectedValues: []float64{},
		},
		{
			name:           "NilWaveform",
			waveformData:   nil,
			expectValid:    false,
			expectedSize:   0,
			expectedValues: nil,
		},
		{
			name:           "LargeWaveform",
			waveformData:   make([]float64, 1000),
			expectValid:    true,
			expectedSize:   1000,
			expectedValues: make([]float64, 1000),
		},
		{
			name:           "ExtremeValues",
			waveformData:   []float64{-1.0, 1.0, 0.0, -0.999, 0.999},
			expectValid:    true,
			expectedSize:   5,
			expectedValues: []float64{-1.0, 1.0, 0.0, -0.999, 0.999},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evp := EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   tt.waveformData,
				Annotations:    []string{"Test waveform"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			}

			err := evp.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSize, len(evp.WaveformData))
				if tt.expectedValues != nil {
					assert.Equal(t, tt.expectedValues, evp.WaveformData)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEVPRecording_TimestampHandling(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timestamp time.Time
		expectErr bool
	}{
		{
			name:      "ValidTimestamp",
			timestamp: baseTime,
			expectErr: false,
		},
		{
			name:      "ZeroTime",
			timestamp: time.Time{},
			expectErr: false,
		},
		{
			name:      "FutureTimestamp",
			timestamp: time.Now().Add(24 * time.Hour),
			expectErr: false,
		},
		{
			name:      "PastTimestamp",
			timestamp: time.Now().Add(-24 * time.Hour),
			expectErr: false,
		},
		{
			name:      "UnixEpoch",
			timestamp: time.Unix(0, 0),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evp := EVPRecording{
				ID:             "evp-123",
				SessionID:      "session-456",
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      tt.timestamp,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"Test timestamp"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      time.Now(),
			}

			err := evp.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.timestamp.Unix(), evp.Timestamp.Unix())
			}
		})
	}
}

func (evp *EVPRecording) Validate() error {
	if evp.ID == "" {
		return fmt.Errorf("EVP recording ID is required")
	}
	if evp.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if evp.FilePath == "" {
		return fmt.Errorf("file path is required")
	}
	if evp.Duration < 0 {
		return fmt.Errorf("duration cannot be negative")
	}
	if evp.WaveformData == nil || len(evp.WaveformData) == 0 {
		return fmt.Errorf("waveform data is required and cannot be empty")
	}
	if evp.DetectionLevel < 0.0 || evp.DetectionLevel > 1.0 {
		return fmt.Errorf("detection level must be between 0.0 and 1.0, got %f", evp.DetectionLevel)
	}
	return nil
}
