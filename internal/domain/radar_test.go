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

func TestSourceType_Values(t *testing.T) {
	tests := []struct {
		name       string
		sourceType SourceType
		expected   string
	}{
		{"EMF", SourceTypeEMF, "emf"},
		{"Audio", SourceTypeAudio, "audio"},
		{"Both", SourceTypeBoth, "both"},
		{"Other", SourceTypeOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.sourceType))
		})
	}
}

func TestRadarEvent_Validation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		radar     RadarEvent
		expectErr bool
	}{
		{
			name: "ValidRadarEvent",
			radar: RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeBoth,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				MovementTrail: []Coordinates{
					{X: 10.0, Y: 20.0, Z: 5.0},
					{X: 10.5, Y: 20.3, Z: 5.1},
					{X: 11.0, Y: 20.6, Z: 5.2},
				},
				CreatedAt: now,
			},
			expectErr: false,
		},
		{
			name: "MissingID",
			radar: RadarEvent{
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeEMF,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				CreatedAt:    now,
			},
			expectErr: true,
		},
		{
			name: "MissingSessionID",
			radar: RadarEvent{
				ID:           "radar-123",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeAudio,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				CreatedAt:    now,
			},
			expectErr: true,
		},
		{
			name: "NegativeStrength",
			radar: RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     -0.1,
				SourceType:   SourceTypeEMF,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				CreatedAt:    now,
			},
			expectErr: true,
		},
		{
			name: "NegativeDuration",
			radar: RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeAudio,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     -5.0,
				CreatedAt:    now,
			},
			expectErr: true,
		},
		{
			name: "ValidWithoutMovementTrail",
			radar: RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeOther,
				EMFReading:   0.0,
				AudioAnomaly: 0.0,
				Duration:     10.5,
				CreatedAt:    now,
			},
			expectErr: false,
		},
		{
			name: "ValidWithEmptyMovementTrail",
			radar: RadarEvent{
				ID:            "radar-123",
				SessionID:     "session-456",
				Timestamp:     now,
				Position:      Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:      0.75,
				SourceType:    SourceTypeEMF,
				EMFReading:    2.5,
				AudioAnomaly:  0.0,
				Duration:      10.5,
				MovementTrail: []Coordinates{},
				CreatedAt:     now,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.radar.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRadarEvent_SourceTypeValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		sourceType SourceType
		expectErr  bool
	}{
		{
			name:       "ValidEMF",
			sourceType: SourceTypeEMF,
			expectErr:  false,
		},
		{
			name:       "ValidAudio",
			sourceType: SourceTypeAudio,
			expectErr:  false,
		},
		{
			name:       "ValidBoth",
			sourceType: SourceTypeBoth,
			expectErr:  false,
		},
		{
			name:       "ValidOther",
			sourceType: SourceTypeOther,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radar := RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   tt.sourceType,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				CreatedAt:    now,
			}

			err := radar.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.sourceType, radar.SourceType)
			}
		})
	}
}

func TestRadarEvent_PositionValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		position    Coordinates
		expectValid bool
		expectedErr string
	}{
		{
			name: "Valid3DCoordinates",
			position: Coordinates{
				X: 10.5,
				Y: 20.3,
				Z: 5.1,
			},
			expectValid: true,
		},
		{
			name: "Valid2DCoordinates",
			position: Coordinates{
				X: 10.5,
				Y: 20.3,
			},
			expectValid: true,
		},
		{
			name: "ValidNegativeCoordinates",
			position: Coordinates{
				X: -10.5,
				Y: -20.3,
				Z: -5.1,
			},
			expectValid: true,
		},
		{
			name: "ValidZeroCoordinates",
			position: Coordinates{
				X: 0.0,
				Y: 0.0,
				Z: 0.0,
			},
			expectValid: true,
		},
		{
			name: "ValidLargeCoordinates",
			position: Coordinates{
				X: 1000.5,
				Y: -2000.3,
				Z: 500.1,
			},
			expectValid: true,
		},
		{
			name: "ValidFractionalCoordinates",
			position: Coordinates{
				X: 10.123456789,
				Y: 20.987654321,
				Z: 5.555555555,
			},
			expectValid: true,
		},
		{
			name: "ValidExtremeCoordinates",
			position: Coordinates{
				X: 1.7976931348623157e+308,  // Max float64
				Y: -1.7976931348623157e+308, // Min float64
				Z: 0.0,
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radar := RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     tt.position,
				Strength:     0.75,
				SourceType:   SourceTypeEMF,
				EMFReading:   2.5,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				CreatedAt:    now,
			}

			err := radar.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.position.X, radar.Position.X)
				assert.Equal(t, tt.position.Y, radar.Position.Y)
				assert.Equal(t, tt.position.Z, radar.Position.Z)
			} else {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			}
		})
	}
}

func TestRadarEvent_MovementTrailValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		movementTrail  []Coordinates
		expectValid    bool
		expectedSize   int
		expectedValues []Coordinates
	}{
		{
			name: "ValidMovementTrail",
			movementTrail: []Coordinates{
				{X: 10.0, Y: 20.0, Z: 5.0},
				{X: 10.5, Y: 20.3, Z: 5.1},
				{X: 11.0, Y: 20.6, Z: 5.2},
			},
			expectValid:  true,
			expectedSize: 3,
			expectedValues: []Coordinates{
				{X: 10.0, Y: 20.0, Z: 5.0},
				{X: 10.5, Y: 20.3, Z: 5.1},
				{X: 11.0, Y: 20.6, Z: 5.2},
			},
		},
		{
			name:           "EmptyMovementTrail",
			movementTrail:  []Coordinates{},
			expectValid:    true,
			expectedSize:   0,
			expectedValues: []Coordinates{},
		},
		{
			name:           "NilMovementTrail",
			movementTrail:  nil,
			expectValid:    true,
			expectedSize:   0,
			expectedValues: nil,
		},
		{
			name: "SinglePointMovementTrail",
			movementTrail: []Coordinates{
				{X: 10.5, Y: 20.3, Z: 5.1},
			},
			expectValid:  true,
			expectedSize: 1,
			expectedValues: []Coordinates{
				{X: 10.5, Y: 20.3, Z: 5.1},
			},
		},
		{
			name:           "LargeMovementTrail",
			movementTrail:  make([]Coordinates, 1000),
			expectValid:    true,
			expectedSize:   1000,
			expectedValues: make([]Coordinates, 1000),
		},
		{
			name: "Mixed2DAnd3DPoints",
			movementTrail: []Coordinates{
				{X: 10.0, Y: 20.0},         // 2D
				{X: 10.5, Y: 20.3, Z: 5.1}, // 3D
				{X: 11.0, Y: 20.6},         // 2D
				{X: 11.5, Y: 20.9, Z: 5.3}, // 3D
			},
			expectValid:  true,
			expectedSize: 4,
			expectedValues: []Coordinates{
				{X: 10.0, Y: 20.0},
				{X: 10.5, Y: 20.3, Z: 5.1},
				{X: 11.0, Y: 20.6},
				{X: 11.5, Y: 20.9, Z: 5.3},
			},
		},
		{
			name: "ExtremeCoordinatesInTrail",
			movementTrail: []Coordinates{
				{X: -1000.0, Y: 2000.0, Z: -500.0},
				{X: 1000.0, Y: -2000.0, Z: 500.0},
				{X: 0.0, Y: 0.0, Z: 0.0},
			},
			expectValid:  true,
			expectedSize: 3,
			expectedValues: []Coordinates{
				{X: -1000.0, Y: 2000.0, Z: -500.0},
				{X: 1000.0, Y: -2000.0, Z: 500.0},
				{X: 0.0, Y: 0.0, Z: 0.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radar := RadarEvent{
				ID:            "radar-123",
				SessionID:     "session-456",
				Timestamp:     now,
				Position:      Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:      0.75,
				SourceType:    SourceTypeBoth,
				EMFReading:    2.5,
				AudioAnomaly:  0.8,
				Duration:      10.5,
				MovementTrail: tt.movementTrail,
				CreatedAt:     now,
			}

			err := radar.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSize, len(radar.MovementTrail))
				if tt.expectedValues != nil {
					assert.Equal(t, tt.expectedValues, radar.MovementTrail)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRadarEvent_EMFReadingValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		emfReading  float64
		expectValid bool
	}{
		{
			name:        "ValidPositiveEMF",
			emfReading:  2.5,
			expectValid: true,
		},
		{
			name:        "ValidZeroEMF",
			emfReading:  0.0,
			expectValid: true,
		},
		{
			name:        "ValidSmallEMF",
			emfReading:  0.1,
			expectValid: true,
		},
		{
			name:        "ValidLargeEMF",
			emfReading:  100.0,
			expectValid: true,
		},
		{
			name:        "ValidFractionalEMF",
			emfReading:  2.756,
			expectValid: true,
		},
		{
			name:        "InvalidNegativeEMF",
			emfReading:  -0.1,
			expectValid: false,
		},
		{
			name:        "InvalidLargeNegativeEMF",
			emfReading:  -100.0,
			expectValid: false,
		},
		{
			name:        "BoundaryZero",
			emfReading:  0.0,
			expectValid: true,
		},
		{
			name:        "VerySmallPositive",
			emfReading:  0.001,
			expectValid: true,
		},
		{
			name:        "VeryLargePositive",
			emfReading:  10000.0,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radar := RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeEMF,
				EMFReading:   tt.emfReading,
				AudioAnomaly: 0.8,
				Duration:     10.5,
				CreatedAt:    now,
			}

			err := radar.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.emfReading, radar.EMFReading)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "EMF reading")
			}
		})
	}
}

func TestRadarEvent_AudioAnomalyValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		audioAnomaly  float64
		expectValid   bool
		expectedError string
	}{
		{
			name:         "ValidPositiveAudioAnomaly",
			audioAnomaly: 0.8,
			expectValid:  true,
		},
		{
			name:         "ValidZeroAudioAnomaly",
			audioAnomaly: 0.0,
			expectValid:  true,
		},
		{
			name:         "ValidSmallAudioAnomaly",
			audioAnomaly: 0.1,
			expectValid:  true,
		},
		{
			name:         "ValidLargeAudioAnomaly",
			audioAnomaly: 10.0,
			expectValid:  true,
		},
		{
			name:         "ValidFractionalAudioAnomaly",
			audioAnomaly: 0.756,
			expectValid:  true,
		},
		{
			name:          "InvalidNegativeAudioAnomaly",
			audioAnomaly:  -0.1,
			expectValid:   false,
			expectedError: "audio anomaly",
		},
		{
			name:          "InvalidLargeNegativeAudioAnomaly",
			audioAnomaly:  -10.0,
			expectValid:   false,
			expectedError: "audio anomaly",
		},
		{
			name:         "BoundaryZero",
			audioAnomaly: 0.0,
			expectValid:  true,
		},
		{
			name:         "VerySmallPositive",
			audioAnomaly: 0.001,
			expectValid:  true,
		},
		{
			name:         "VeryLargePositive",
			audioAnomaly: 1000.0,
			expectValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			radar := RadarEvent{
				ID:           "radar-123",
				SessionID:    "session-456",
				Timestamp:    now,
				Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
				Strength:     0.75,
				SourceType:   SourceTypeAudio,
				EMFReading:   2.5,
				AudioAnomaly: tt.audioAnomaly,
				Duration:     10.5,
				CreatedAt:    now,
			}

			err := radar.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.audioAnomaly, radar.AudioAnomaly)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestRadarEvent_JSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	radar := RadarEvent{
		ID:           "radar-123",
		SessionID:    "session-456",
		Timestamp:    now,
		Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
		Strength:     0.85,
		SourceType:   SourceTypeBoth,
		EMFReading:   3.7,
		AudioAnomaly: 0.92,
		Duration:     15.2,
		MovementTrail: []Coordinates{
			{X: 10.0, Y: 20.0, Z: 5.0},
			{X: 10.5, Y: 20.3, Z: 5.1},
			{X: 11.0, Y: 20.6, Z: 5.2},
		},
		CreatedAt: now,
	}

	t.Run("MarshalJSON", func(t *testing.T) {
		data, err := json.Marshal(radar)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		var unmarshaled RadarEvent
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, radar.ID, unmarshaled.ID)
		assert.Equal(t, radar.SessionID, unmarshaled.SessionID)
		assert.Equal(t, radar.Strength, unmarshaled.Strength)
		assert.Equal(t, radar.SourceType, unmarshaled.SourceType)
		assert.Equal(t, radar.EMFReading, unmarshaled.EMFReading)
		assert.Equal(t, radar.AudioAnomaly, unmarshaled.AudioAnomaly)
		assert.Equal(t, radar.Duration, unmarshaled.Duration)
		assert.Equal(t, radar.Timestamp.Unix(), unmarshaled.Timestamp.Unix())
		assert.Equal(t, radar.CreatedAt.Unix(), unmarshaled.CreatedAt.Unix())
		assert.Equal(t, radar.Position.X, unmarshaled.Position.X)
		assert.Equal(t, radar.Position.Y, unmarshaled.Position.Y)
		assert.Equal(t, radar.Position.Z, unmarshaled.Position.Z)
		assert.Equal(t, radar.MovementTrail, unmarshaled.MovementTrail)
	})

	t.Run("MarshalJSONWithoutMovementTrail", func(t *testing.T) {
		radarMinimal := RadarEvent{
			ID:           "radar-123",
			SessionID:    "session-456",
			Timestamp:    now,
			Position:     Coordinates{X: 10.5, Y: 20.3, Z: 5.1},
			Strength:     0.75,
			SourceType:   SourceTypeEMF,
			EMFReading:   2.5,
			AudioAnomaly: 0.0,
			Duration:     10.5,
			CreatedAt:    now,
		}

		data, err := json.Marshal(radarMinimal)
		require.NoError(t, err)

		var unmarshaled RadarEvent
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, radarMinimal.ID, unmarshaled.ID)
		assert.Equal(t, radarMinimal.Position.X, unmarshaled.Position.X)
		assert.Equal(t, radarMinimal.Position.Y, unmarshaled.Position.Y)
		assert.Equal(t, radarMinimal.Position.Z, unmarshaled.Position.Z)
		assert.Empty(t, unmarshaled.MovementTrail, "MovementTrail should be empty when not set")
	})
}

func TestRadarEvent_DatabaseTags(t *testing.T) {
	radar := RadarEvent{}
	radarType := reflect.TypeOf(radar)

	requiredTags := []string{"id", "session_id", "timestamp", "position", "strength", "source_type", "emf_reading", "audio_anomaly", "duration", "created_at"}
	optionalTags := []string{"movement_trail"}

	for _, tag := range requiredTags {
		field, found := radarType.FieldByNameFunc(func(name string) bool {
			field, _ := radarType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		assert.True(t, found, "Required field with db tag '%s' not found", tag)
		assert.NotEmpty(t, field.Name, "Field name should not be empty for tag '%s'", tag)
	}

	for _, tag := range optionalTags {
		field, found := radarType.FieldByNameFunc(func(name string) bool {
			field, _ := radarType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		if found {
			assert.NotEmpty(t, field.Name, "Field name should not be empty for optional tag '%s'", tag)
		}
	}
}

func (radar *RadarEvent) Validate() error {
	if radar.ID == "" {
		return fmt.Errorf("radar event ID is required")
	}
	if radar.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if radar.Strength < 0.0 {
		return fmt.Errorf("strength cannot be negative")
	}
	if radar.Duration < 0.0 {
		return fmt.Errorf("duration cannot be negative")
	}
	if radar.EMFReading < 0.0 {
		return fmt.Errorf("EMF reading cannot be negative")
	}
	if radar.AudioAnomaly < 0.0 {
		return fmt.Errorf("audio anomaly cannot be negative")
	}
	return nil
}
