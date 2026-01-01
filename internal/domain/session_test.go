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

func TestSessionStatus_Values(t *testing.T) {
	tests := []struct {
		name     string
		status   SessionStatus
		expected string
	}{
		{"Active", SessionStatusActive, "active"},
		{"Paused", SessionStatusPaused, "paused"},
		{"Complete", SessionStatusComplete, "complete"},
		{"Archived", SessionStatusArchived, "archived"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestSessionStatus_Transitions(t *testing.T) {
	tests := []struct {
		name     string
		from     SessionStatus
		to       SessionStatus
		expected bool
	}{
		{"ActiveToPaused", SessionStatusActive, SessionStatusPaused, true},
		{"PausedToActive", SessionStatusPaused, SessionStatusActive, true},
		{"ActiveToComplete", SessionStatusActive, SessionStatusComplete, true},
		{"CompleteToArchived", SessionStatusComplete, SessionStatusArchived, true},
		{"ArchivedToActive", SessionStatusArchived, SessionStatusActive, false},
		{"CompleteToActive", SessionStatusComplete, SessionStatusActive, false},
		{"ArchivedToPaused", SessionStatusArchived, SessionStatusPaused, false},
		{"ActiveToActive", SessionStatusActive, SessionStatusActive, true},
		{"PausedToPaused", SessionStatusPaused, SessionStatusPaused, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSessionStatusTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLocation_Validation(t *testing.T) {
	tests := []struct {
		name      string
		location  Location
		expectErr bool
	}{
		{
			name: "ValidLocation",
			location: Location{
				Latitude:    40.7128,
				Longitude:   -74.0060,
				Address:     "123 Test St",
				Description: "Test location",
				Venue:       "Test Venue",
			},
			expectErr: false,
		},
		{
			name: "InvalidLatitudeTooHigh",
			location: Location{
				Latitude:    91.0,
				Longitude:   -74.0060,
				Address:     "123 Test St",
				Description: "Test location",
				Venue:       "Test Venue",
			},
			expectErr: true,
		},
		{
			name: "InvalidLatitudeTooLow",
			location: Location{
				Latitude:    -91.0,
				Longitude:   -74.0060,
				Address:     "123 Test St",
				Description: "Test location",
				Venue:       "Test Venue",
			},
			expectErr: true,
		},
		{
			name: "InvalidLongitudeTooHigh",
			location: Location{
				Latitude:    40.7128,
				Longitude:   181.0,
				Address:     "123 Test St",
				Description: "Test location",
				Venue:       "Test Venue",
			},
			expectErr: true,
		},
		{
			name: "InvalidLongitudeTooLow",
			location: Location{
				Latitude:    40.7128,
				Longitude:   -181.0,
				Address:     "123 Test St",
				Description: "Test location",
				Venue:       "Test Venue",
			},
			expectErr: true,
		},
		{
			name: "ValidBoundaryLatitude",
			location: Location{
				Latitude:    90.0,
				Longitude:   0.0,
				Address:     "North Pole",
				Description: "Arctic location",
				Venue:       "Research Station",
			},
			expectErr: false,
		},
		{
			name: "ValidBoundaryLongitude",
			location: Location{
				Latitude:    0.0,
				Longitude:   180.0,
				Address:     "International Date Line",
				Description: "Pacific location",
				Venue:       "Island",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.location.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnvironmental_Validation(t *testing.T) {
	tests := []struct {
		name          string
		env           Environmental
		expectErr     bool
		expectedField string
	}{
		{
			name: "ValidEnvironmental",
			env: Environmental{
				Temperature: 20.5,
				Humidity:    45.0,
				Pressure:    1013.25,
				EMFLevel:    0.1,
				LightLevel:  50.0,
				NoiseLevel:  30.0,
			},
			expectErr: false,
		},
		{
			name: "InvalidTemperatureTooLow",
			env: Environmental{
				Temperature: -274.0,
				Humidity:    45.0,
				Pressure:    1013.25,
				EMFLevel:    0.1,
				LightLevel:  50.0,
				NoiseLevel:  30.0,
			},
			expectErr:     true,
			expectedField: "temperature",
		},
		{
			name: "InvalidHumidityTooLow",
			env: Environmental{
				Temperature: 20.5,
				Humidity:    -5.0,
				Pressure:    1013.25,
				EMFLevel:    0.1,
				LightLevel:  50.0,
				NoiseLevel:  30.0,
			},
			expectErr:     true,
			expectedField: "humidity",
		},
		{
			name: "InvalidHumidityTooHigh",
			env: Environmental{
				Temperature: 20.5,
				Humidity:    105.0,
				Pressure:    1013.25,
				EMFLevel:    0.1,
				LightLevel:  50.0,
				NoiseLevel:  30.0,
			},
			expectErr:     true,
			expectedField: "humidity",
		},
		{
			name: "InvalidPressureTooLow",
			env: Environmental{
				Temperature: 20.5,
				Humidity:    45.0,
				Pressure:    100.0,
				EMFLevel:    0.1,
				LightLevel:  50.0,
				NoiseLevel:  30.0,
			},
			expectErr:     true,
			expectedField: "pressure",
		},
		{
			name: "InvalidEMFLevelNegative",
			env: Environmental{
				Temperature: 20.5,
				Humidity:    45.0,
				Pressure:    1013.25,
				EMFLevel:    -0.1,
				LightLevel:  50.0,
				NoiseLevel:  30.0,
			},
			expectErr:     true,
			expectedField: "EMF level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedField != "" {
					assert.Contains(t, err.Error(), tt.expectedField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSession_Validation(t *testing.T) {
	now := time.Now()
	validLocation := Location{
		Latitude:    40.7128,
		Longitude:   -74.0060,
		Address:     "123 Test St",
		Description: "Test location",
		Venue:       "Test Venue",
	}
	validEnvironmental := Environmental{
		Temperature: 20.5,
		Humidity:    45.0,
		Pressure:    1013.25,
		EMFLevel:    0.1,
		LightLevel:  50.0,
		NoiseLevel:  30.0,
	}

	tests := []struct {
		name      string
		session   Session
		expectErr bool
	}{
		{
			name: "ValidSession",
			session: Session{
				ID:            "test-session-id",
				Title:         "Test Session",
				Location:      validLocation,
				StartTime:     now,
				Notes:         "Test notes",
				Environmental: validEnvironmental,
				Status:        SessionStatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: false,
		},
		{
			name: "MissingID",
			session: Session{
				Title:         "Test Session",
				Location:      validLocation,
				StartTime:     now,
				Notes:         "Test notes",
				Environmental: validEnvironmental,
				Status:        SessionStatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: true,
		},
		{
			name: "EmptyTitle",
			session: Session{
				ID:            "test-session-id",
				Title:         "",
				Location:      validLocation,
				StartTime:     now,
				Notes:         "Test notes",
				Environmental: validEnvironmental,
				Status:        SessionStatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: true,
		},
		{
			name: "InvalidLocation",
			session: Session{
				ID:            "test-session-id",
				Title:         "Test Session",
				Location:      Location{Latitude: 91.0, Longitude: 0.0},
				StartTime:     now,
				Notes:         "Test notes",
				Environmental: validEnvironmental,
				Status:        SessionStatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: true,
		},
		{
			name: "InvalidEnvironmental",
			session: Session{
				ID:            "test-session-id",
				Title:         "Test Session",
				Location:      validLocation,
				StartTime:     now,
				Notes:         "Test notes",
				Environmental: Environmental{Temperature: -300.0, Humidity: 45.0},
				Status:        SessionStatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: true,
		},
		{
			name: "EndTimeBeforeStartTime",
			session: Session{
				ID:            "test-session-id",
				Title:         "Test Session",
				Location:      validLocation,
				StartTime:     now,
				EndTime:       &[]time.Time{now.Add(-1 * time.Hour)}[0],
				Notes:         "Test notes",
				Environmental: validEnvironmental,
				Status:        SessionStatusActive,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: true,
		},
		{
			name: "ValidEndTimeAfterStartTime",
			session: Session{
				ID:            "test-session-id",
				Title:         "Test Session",
				Location:      validLocation,
				StartTime:     now,
				EndTime:       &[]time.Time{now.Add(1 * time.Hour)}[0],
				Notes:         "Test notes",
				Environmental: validEnvironmental,
				Status:        SessionStatusComplete,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSession_JSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	session := Session{
		ID:    "test-session-id",
		Title: "Test Session",
		Location: Location{
			Latitude:    40.7128,
			Longitude:   -74.0060,
			Address:     "123 Test St",
			Description: "Test location",
			Venue:       "Test Venue",
		},
		StartTime: now,
		Notes:     "Test notes",
		Environmental: Environmental{
			Temperature: 20.5,
			Humidity:    45.0,
			Pressure:    1013.25,
			EMFLevel:    0.1,
			LightLevel:  50.0,
			NoiseLevel:  30.0,
		},
		Status:    SessionStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("MarshalJSON", func(t *testing.T) {
		data, err := json.Marshal(session)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		var unmarshaled Session
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, session.ID, unmarshaled.ID)
		assert.Equal(t, session.Title, unmarshaled.Title)
		assert.Equal(t, session.Status, unmarshaled.Status)
		assert.Equal(t, session.Location.Latitude, unmarshaled.Location.Latitude)
		assert.Equal(t, session.Location.Longitude, unmarshaled.Location.Longitude)
		assert.Equal(t, session.Environmental.Temperature, unmarshaled.Environmental.Temperature)
	})

	t.Run("MarshalJSONWithEndTime", func(t *testing.T) {
		endTime := now.Add(1 * time.Hour)
		session.EndTime = &endTime

		data, err := json.Marshal(session)
		require.NoError(t, err)

		var unmarshaled Session
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		require.NotNil(t, unmarshaled.EndTime)
		assert.Equal(t, endTime.Unix(), unmarshaled.EndTime.Unix())
	})

	t.Run("MarshalJSONWithCollections", func(t *testing.T) {
		session.EVPRecordings = []EVPRecording{
			{
				ID:             "evp-1",
				SessionID:      session.ID,
				FilePath:       "/path/to/recording.wav",
				Duration:       30.5,
				Timestamp:      now,
				WaveformData:   []float64{0.1, 0.2, 0.3},
				Annotations:    []string{"test annotation"},
				Quality:        EVPQualityGood,
				DetectionLevel: 0.75,
				CreatedAt:      now,
			},
		}

		data, err := json.Marshal(session)
		require.NoError(t, err)

		var unmarshaled Session
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Len(t, unmarshaled.EVPRecordings, 1)
		assert.Equal(t, "evp-1", unmarshaled.EVPRecordings[0].ID)
	})
}

func TestSession_DatabaseTags(t *testing.T) {
	session := Session{}

	sessionType := reflect.TypeOf(session)

	requiredTags := []string{"id", "title", "location", "start_time", "notes", "environmental", "status", "created_at", "updated_at"}
	optionalTags := []string{"end_time"}

	for _, tag := range requiredTags {
		field, found := sessionType.FieldByNameFunc(func(name string) bool {
			field, _ := sessionType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		assert.True(t, found, "Required field with db tag '%s' not found", tag)
		assert.NotEmpty(t, field.Name, "Field name should not be empty for tag '%s'", tag)
	}

	for _, tag := range optionalTags {
		field, found := sessionType.FieldByNameFunc(func(name string) bool {
			field, _ := sessionType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		if found {
			assert.NotEmpty(t, field.Name, "Field name should not be empty for optional tag '%s'", tag)
		}
	}
}

func TestSession_Creation(t *testing.T) {
	now := time.Now()
	validLocation := Location{
		Latitude:    40.7128,
		Longitude:   -74.0060,
		Address:     "123 Test St",
		Description: "Test location",
		Venue:       "Test Venue",
	}
	validEnvironmental := Environmental{
		Temperature: 20.5,
		Humidity:    45.0,
		Pressure:    1013.25,
		EMFLevel:    0.1,
		LightLevel:  50.0,
		NoiseLevel:  30.0,
	}

	t.Run("ValidSessionCreation", func(t *testing.T) {
		session := &Session{
			ID:            "test-session-id",
			Title:         "Test Session",
			Location:      validLocation,
			StartTime:     now,
			Notes:         "Test notes",
			Environmental: validEnvironmental,
			Status:        SessionStatusActive,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		err := session.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "test-session-id", session.ID)
		assert.Equal(t, "Test Session", session.Title)
		assert.Equal(t, SessionStatusActive, session.Status)
	})

	t.Run("SessionWithAllRequiredFields", func(t *testing.T) {
		session := &Session{
			ID:            "test-2",
			Title:         "Complete Test Session",
			Location:      validLocation,
			StartTime:     now,
			Environmental: validEnvironmental,
			Status:        SessionStatusPaused,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		err := session.Validate()
		assert.NoError(t, err)
		assert.Equal(t, "test-2", session.ID)
		assert.Equal(t, SessionStatusPaused, session.Status)
	})
}

func isValidSessionStatusTransition(from, to SessionStatus) bool {
	if from == to {
		return true
	}

	switch from {
	case SessionStatusActive:
		return to == SessionStatusPaused || to == SessionStatusComplete
	case SessionStatusPaused:
		return to == SessionStatusActive || to == SessionStatusComplete
	case SessionStatusComplete:
		return to == SessionStatusArchived
	case SessionStatusArchived:
		return false
	default:
		return false
	}
}

func (l Location) Validate() error {
	if l.Latitude < -90.0 || l.Latitude > 90.0 {
		return fmt.Errorf("latitude must be between -90 and 90 degrees, got %f", l.Latitude)
	}
	if l.Longitude < -180.0 || l.Longitude > 180.0 {
		return fmt.Errorf("longitude must be between -180 and 180 degrees, got %f", l.Longitude)
	}
	return nil
}

func (e Environmental) Validate() error {
	if e.Temperature < -273.15 {
		return fmt.Errorf("temperature cannot be below absolute zero, got %f", e.Temperature)
	}
	if e.Humidity < 0.0 || e.Humidity > 100.0 {
		return fmt.Errorf("humidity must be between 0 and 100 percent, got %f", e.Humidity)
	}
	if e.Pressure < 500.0 || e.Pressure > 1200.0 {
		return fmt.Errorf("pressure must be within reasonable atmospheric range, got %f", e.Pressure)
	}
	if e.EMFLevel < 0.0 {
		return fmt.Errorf("EMF level cannot be negative, got %f", e.EMFLevel)
	}
	if e.LightLevel < 0.0 {
		return fmt.Errorf("light level cannot be negative, got %f", e.LightLevel)
	}
	if e.NoiseLevel < 0.0 {
		return fmt.Errorf("noise level cannot be negative, got %f", e.NoiseLevel)
	}
	return nil
}

func (s *Session) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("session ID is required")
	}
	if s.Title == "" {
		return fmt.Errorf("session title is required")
	}
	if err := s.Location.Validate(); err != nil {
		return fmt.Errorf("invalid location: %w", err)
	}
	if err := s.Environmental.Validate(); err != nil {
		return fmt.Errorf("invalid environmental data: %w", err)
	}
	if s.EndTime != nil && s.EndTime.Before(s.StartTime) {
		return fmt.Errorf("end time cannot be before start time")
	}
	return nil
}
