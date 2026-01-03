package domain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVOXEvent_Validation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		vox       VOXEvent
		expectErr bool
	}{
		{
			name: "ValidVOXEvent",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				UserResponse:    "Yes, I can hear you",
				ResponseDelay:   2.5,
				CreatedAt:       now,
			},
			expectErr: false,
		},
		{
			name: "MissingID",
			vox: VOXEvent{
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "MissingSessionID",
			vox: VOXEvent{
				ID:              "vox-123",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "EmptyGeneratedText",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "EmptyPhoneticBank",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "TriggerStrengthTooLow",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: -0.1,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "TriggerStrengthTooHigh",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 1.1,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "EmptyLanguagePack",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "ValidWithoutUserResponse",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: false,
		},
		{
			name: "ValidWithNegativeResponseDelay",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				UserResponse:    "Response",
				ResponseDelay:   -1.0,
				CreatedAt:       now,
			},
			expectErr: true,
		},
		{
			name: "BoundaryTriggerStrengthZero",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.0,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: false,
		},
		{
			name: "BoundaryTriggerStrengthOne",
			vox: VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 1.0,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vox.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVOXEvent_TriggerStrengthValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		triggerStrength  float64
		expectErr        bool
		expectedErrField string
	}{
		{
			name:            "ValidMidRange",
			triggerStrength: 0.5,
			expectErr:       false,
		},
		{
			name:            "ValidLowerBoundary",
			triggerStrength: 0.0,
			expectErr:       false,
		},
		{
			name:            "ValidUpperBoundary",
			triggerStrength: 1.0,
			expectErr:       false,
		},
		{
			name:             "InvalidNegative",
			triggerStrength:  -0.01,
			expectErr:        true,
			expectedErrField: "trigger strength",
		},
		{
			name:             "InvalidAboveOne",
			triggerStrength:  1.01,
			expectErr:        true,
			expectedErrField: "trigger strength",
		},
		{
			name:             "InvalidLargeNegative",
			triggerStrength:  -100.0,
			expectErr:        true,
			expectedErrField: "trigger strength",
		},
		{
			name:             "InvalidLargePositive",
			triggerStrength:  100.0,
			expectErr:        true,
			expectedErrField: "trigger strength",
		},
		{
			name:            "ValidVerySmallPositive",
			triggerStrength: 0.001,
			expectErr:       false,
		},
		{
			name:            "ValidVeryCloseToOne",
			triggerStrength: 0.999,
			expectErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vox := VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Test message",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: tt.triggerStrength,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			}

			err := vox.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				if tt.expectedErrField != "" {
					assert.Contains(t, err.Error(), tt.expectedErrField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVOXEvent_LanguagePackValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		languagePack  string
		expectValid   bool
		expectedError string
	}{
		{
			name:         "ValidEnglishUS",
			languagePack: "en-US",
			expectValid:  true,
		},
		{
			name:         "ValidEnglishGB",
			languagePack: "en-GB",
			expectValid:  true,
		},
		{
			name:         "ValidSpanish",
			languagePack: "es-ES",
			expectValid:  true,
		},
		{
			name:         "ValidFrench",
			languagePack: "fr-FR",
			expectValid:  true,
		},
		{
			name:          "EmptyLanguagePack",
			languagePack:  "",
			expectValid:   false,
			expectedError: "language pack is required",
		},
		{
			name:          "WhitespaceOnly",
			languagePack:  "   ",
			expectValid:   false,
			expectedError: "language pack is required",
		},
		{
			name:         "ValidGerman",
			languagePack: "de-DE",
			expectValid:  true,
		},
		{
			name:         "ValidItalian",
			languagePack: "it-IT",
			expectValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vox := VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Test message",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    tt.languagePack,
				ModulationType:  "amplitude",
				CreatedAt:       now,
			}

			err := vox.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.languagePack, vox.LanguagePack)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestVOXEvent_PhoneticBankValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		phoneticBank  string
		expectValid   bool
		expectedError string
	}{
		{
			name:         "ValidEnglishStandard",
			phoneticBank: "english_standard",
			expectValid:  true,
		},
		{
			name:         "ValidEnglishEnhanced",
			phoneticBank: "english_enhanced",
			expectValid:  true,
		},
		{
			name:         "ValidSpanishBasic",
			phoneticBank: "spanish_basic",
			expectValid:  true,
		},
		{
			name:         "ValidCustom",
			phoneticBank: "custom_paranormal",
			expectValid:  true,
		},
		{
			name:          "EmptyPhoneticBank",
			phoneticBank:  "",
			expectValid:   false,
			expectedError: "phonetic bank is required",
		},
		{
			name:          "WhitespaceOnly",
			phoneticBank:  "   ",
			expectValid:   false,
			expectedError: "phonetic bank is required",
		},
		{
			name:         "SingleCharacter",
			phoneticBank: "a",
			expectValid:  true,
		},
		{
			name:         "LongName",
			phoneticBank: "very_long_and_detailed_phonetic_bank_name_for_specialized_paranormal_investigation",
			expectValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vox := VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Test message",
				PhoneticBank:    tt.phoneticBank,
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			}

			err := vox.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.phoneticBank, vox.PhoneticBank)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestVOXEvent_FrequencyDataValidation(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		frequencyData  []float64
		expectValid    bool
		expectedSize   int
		expectedValues []float64
	}{
		{
			name:           "ValidFrequencyData",
			frequencyData:  []float64{100.0, 200.0, 300.0, 400.0, 500.0},
			expectValid:    true,
			expectedSize:   5,
			expectedValues: []float64{100.0, 200.0, 300.0, 400.0, 500.0},
		},
		{
			name:           "SingleFrequency",
			frequencyData:  []float64{440.0},
			expectValid:    true,
			expectedSize:   1,
			expectedValues: []float64{440.0},
		},
		{
			name:           "EmptyFrequencyData",
			frequencyData:  []float64{},
			expectValid:    true,
			expectedSize:   0,
			expectedValues: []float64{},
		},
		{
			name:           "NilFrequencyData",
			frequencyData:  nil,
			expectValid:    true,
			expectedSize:   0,
			expectedValues: nil,
		},
		{
			name:           "LargeFrequencyDataset",
			frequencyData:  make([]float64, 1000),
			expectValid:    true,
			expectedSize:   1000,
			expectedValues: make([]float64, 1000),
		},
		{
			name:           "ExtremeFrequencies",
			frequencyData:  []float64{20.0, 20000.0, 1000.0},
			expectValid:    true,
			expectedSize:   3,
			expectedValues: []float64{20.0, 20000.0, 1000.0},
		},
		{
			name:           "ZeroFrequencies",
			frequencyData:  []float64{0.0, 0.0, 0.0},
			expectValid:    true,
			expectedSize:   3,
			expectedValues: []float64{0.0, 0.0, 0.0},
		},
		{
			name:           "NegativeFrequencies",
			frequencyData:  []float64{-100.0, 100.0},
			expectValid:    true,
			expectedSize:   2,
			expectedValues: []float64{-100.0, 100.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vox := VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Test message",
				PhoneticBank:    "english_standard",
				FrequencyData:   tt.frequencyData,
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				CreatedAt:       now,
			}

			err := vox.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSize, len(vox.FrequencyData))
				if tt.expectedValues != nil {
					assert.Equal(t, tt.expectedValues, vox.FrequencyData)
				}
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestVOXEvent_ResponseDelayCalculations(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		userResponse  string
		responseDelay float64
		expectValid   bool
		expectedError string
	}{
		{
			name:          "ValidWithUserResponseAndDelay",
			userResponse:  "Yes, I can hear you",
			responseDelay: 2.5,
			expectValid:   true,
		},
		{
			name:          "ValidWithUserResponseNoDelay",
			userResponse:  "Response",
			responseDelay: 0.0,
			expectValid:   true,
		},
		{
			name:          "ValidWithoutUserResponseWithoutDelay",
			userResponse:  "",
			responseDelay: 0.0,
			expectValid:   true,
		},
		{
			name:          "InvalidNegativeDelay",
			userResponse:  "Response",
			responseDelay: -1.0,
			expectValid:   false,
			expectedError: "response delay cannot be negative",
		},
		{
			name:          "InvalidUserResponseWithoutDelay",
			userResponse:  "Response",
			responseDelay: 0.0,
			expectValid:   true, // No delay is valid
		},
		{
			name:          "ValidVeryShortDelay",
			userResponse:  "Quick response",
			responseDelay: 0.1,
			expectValid:   true,
		},
		{
			name:          "ValidVeryLongDelay",
			userResponse:  "Slow response",
			responseDelay: 300.0,
			expectValid:   true,
		},
		{
			name:          "ValidEmptyUserResponseWithZeroDelay",
			userResponse:  "",
			responseDelay: 0.0,
			expectValid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vox := VOXEvent{
				ID:              "vox-123",
				SessionID:       "session-456",
				Timestamp:       now,
				GeneratedText:   "Hello, is anyone there?",
				PhoneticBank:    "english_standard",
				FrequencyData:   []float64{100.0, 200.0, 300.0},
				TriggerStrength: 0.75,
				LanguagePack:    "en-US",
				ModulationType:  "amplitude",
				UserResponse:    tt.userResponse,
				ResponseDelay:   tt.responseDelay,
				CreatedAt:       now,
			}

			err := vox.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if tt.expectedError != "" {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestVOXEvent_JSONSerialization(t *testing.T) {
	now := time.Now().UTC()
	vox := VOXEvent{
		ID:              "vox-123",
		SessionID:       "session-456",
		Timestamp:       now,
		GeneratedText:   "Hello, is anyone there?",
		PhoneticBank:    "english_standard",
		FrequencyData:   []float64{100.0, 200.0, 300.0, 400.0, 500.0},
		TriggerStrength: 0.85,
		LanguagePack:    "en-US",
		ModulationType:  "amplitude",
		UserResponse:    "Yes, I can hear you",
		ResponseDelay:   2.5,
		CreatedAt:       now,
	}

	t.Run("MarshalJSON", func(t *testing.T) {
		data, err := json.Marshal(vox)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		var unmarshaled VOXEvent
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, vox.ID, unmarshaled.ID)
		assert.Equal(t, vox.SessionID, unmarshaled.SessionID)
		assert.Equal(t, vox.GeneratedText, unmarshaled.GeneratedText)
		assert.Equal(t, vox.PhoneticBank, unmarshaled.PhoneticBank)
		assert.Equal(t, vox.TriggerStrength, unmarshaled.TriggerStrength)
		assert.Equal(t, vox.LanguagePack, unmarshaled.LanguagePack)
		assert.Equal(t, vox.ModulationType, unmarshaled.ModulationType)
		assert.Equal(t, vox.UserResponse, unmarshaled.UserResponse)
		assert.Equal(t, vox.ResponseDelay, unmarshaled.ResponseDelay)
		assert.Equal(t, vox.Timestamp.Unix(), unmarshaled.Timestamp.Unix())
		assert.Equal(t, vox.CreatedAt.Unix(), unmarshaled.CreatedAt.Unix())
		assert.Equal(t, vox.FrequencyData, unmarshaled.FrequencyData)
	})

	t.Run("MarshalJSONWithoutOptionalFields", func(t *testing.T) {
		voxMinimal := VOXEvent{
			ID:              "vox-123",
			SessionID:       "session-456",
			Timestamp:       now,
			GeneratedText:   "Basic message",
			PhoneticBank:    "english_standard",
			FrequencyData:   []float64{100.0, 200.0, 300.0},
			TriggerStrength: 0.75,
			LanguagePack:    "en-US",
			ModulationType:  "frequency",
			CreatedAt:       now,
		}

		data, err := json.Marshal(voxMinimal)
		require.NoError(t, err)

		var unmarshaled VOXEvent
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, voxMinimal.ID, unmarshaled.ID)
		assert.Empty(t, unmarshaled.UserResponse, "UserResponse should be empty when not set")
		assert.Equal(t, 0.0, unmarshaled.ResponseDelay, "ResponseDelay should be zero when not set")
	})
}

func TestVOXEvent_DatabaseTags(t *testing.T) {
	vox := VOXEvent{}
	voxType := reflect.TypeOf(vox)

	requiredTags := []string{"id", "session_id", "timestamp", "generated_text", "phonetic_bank", "frequency_data", "trigger_strength", "language_pack", "modulation_type", "created_at"}
	optionalTags := []string{"user_response", "response_delay"}

	for _, tag := range requiredTags {
		field, found := voxType.FieldByNameFunc(func(name string) bool {
			field, _ := voxType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		assert.True(t, found, "Required field with db tag '%s' not found", tag)
		assert.NotEmpty(t, field.Name, "Field name should not be empty for tag '%s'", tag)
	}

	for _, tag := range optionalTags {
		field, found := voxType.FieldByNameFunc(func(name string) bool {
			field, _ := voxType.FieldByName(name)
			dbTag := field.Tag.Get("db")
			return dbTag == tag
		})
		if found {
			assert.NotEmpty(t, field.Name, "Field name should not be empty for optional tag '%s'", tag)
		}
	}
}

func (vox *VOXEvent) Validate() error {
	if vox.ID == "" {
		return fmt.Errorf("VOX event ID is required")
	}
	if vox.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if strings.TrimSpace(vox.GeneratedText) == "" {
		return fmt.Errorf("generated text is required")
	}
	if strings.TrimSpace(vox.PhoneticBank) == "" {
		return fmt.Errorf("phonetic bank is required")
	}
	if vox.TriggerStrength < 0.0 || vox.TriggerStrength > 1.0 {
		return fmt.Errorf("trigger strength must be between 0.0 and 1.0, got %f", vox.TriggerStrength)
	}
	if strings.TrimSpace(vox.LanguagePack) == "" {
		return fmt.Errorf("language pack is required")
	}
	if vox.ResponseDelay < 0.0 {
		return fmt.Errorf("response delay cannot be negative")
	}
	return nil
}
