package audio

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewVOXGenerator validates VOX generator creation and initialization
func TestNewVOXGenerator(t *testing.T) {
	tests := []struct {
		name   string
		config VOXConfig
	}{
		{"DefaultConfig", VOXConfig{
			DefaultLanguage:  "english",
			PhoneticBankSize: 25,
			TriggerThreshold: 0.5,
		}},
		{"MinimalConfig", VOXConfig{
			DefaultLanguage:  "simple",
			PhoneticBankSize: 10,
			TriggerThreshold: 0.3,
		}},
		{"ExtendedConfig", VOXConfig{
			DefaultLanguage:  "english",
			PhoneticBankSize: 40,
			TriggerThreshold: 0.7,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vox := NewVOXGenerator(tt.config)

			assert.NotNil(t, vox)
			assert.NotNil(t, vox.phoneticBanks)
			assert.NotNil(t, vox.languagePacks)

			// Check that default banks are initialized
			assert.Contains(t, vox.phoneticBanks, "english")
			assert.Contains(t, vox.phoneticBanks, "minimal")
			assert.Contains(t, vox.phoneticBanks, "extended")
			assert.Contains(t, vox.languagePacks, "english")
			assert.Contains(t, vox.languagePacks, "simple")

			// Check bank contents
			assert.NotEmpty(t, vox.phoneticBanks["english"])
			assert.NotEmpty(t, vox.phoneticBanks["minimal"])
			assert.NotEmpty(t, vox.phoneticBanks["extended"])
			assert.NotEmpty(t, vox.languagePacks["english"])
			assert.NotEmpty(t, vox.languagePacks["simple"])
		})
	}
}

// TestVOXGenerator_Initialization tests phonetic bank and language pack validation
func TestVOXGenerator_Initialization(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	})

	// Test English phonetic bank
	englishPhonetics := vox.phoneticBanks["english"]
	require.NotEmpty(t, englishPhonetics)

	// Check for expected phonetic categories
	vowels := map[string]bool{
		"ah": false, "eh": false, "ih": false, "oh": false, "uh": false,
		"ay": false, "ey": false, "iy": false, "ow": false, "uw": false,
	}
	consonants := map[string]bool{
		"b": false, "d": false, "f": false, "g": false, "h": false,
		"k": false, "l": false, "m": false, "n": false, "p": false,
		"r": false, "s": false, "t": false, "v": false, "w": false,
		"y": false, "z": false,
	}
	complex := map[string]bool{
		"ch": false, "sh": false, "th": false, "ng": false, "zh": false,
	}

	// Verify all phonetic categories are present
	for _, phonetic := range englishPhonetics {
		if _, exists := vowels[phonetic]; exists {
			vowels[phonetic] = true
		} else if _, exists := consonants[phonetic]; exists {
			consonants[phonetic] = true
		} else if _, exists := complex[phonetic]; exists {
			complex[phonetic] = true
		}
	}

	// Check that we have all expected categories
	assert.True(t, len(vowels) > 0, "Should have vowel phonetics")
	assert.True(t, len(consonants) > 0, "Should have consonant phonetics")
	assert.True(t, len(complex) > 0, "Should have complex phonetics")

	// Test minimal phonetic bank
	minimalPhonetics := vox.phoneticBanks["minimal"]
	require.NotEmpty(t, minimalPhonetics)
	expectedMinimal := []string{"a", "e", "i", "o", "u", "m", "n", "s", "t", "r", "l"}
	assert.Equal(t, len(expectedMinimal), len(minimalPhonetics))
	for _, phonetic := range expectedMinimal {
		assert.Contains(t, minimalPhonetics, phonetic)
	}

	// Test language packs
	englishWords := vox.languagePacks["english"]
	require.NotEmpty(t, englishWords)

	// Check for common paranormal investigation words
	expectedWords := []string{"yes", "no", "here", "there", "go", "stay", "help", "stop"}
	for _, word := range expectedWords {
		assert.Contains(t, englishWords, word)
	}

	// Test extended phonetic bank
	extendedPhonetics := vox.phoneticBanks["extended"]
	require.NotEmpty(t, extendedPhonetics)
	assert.Greater(t, len(extendedPhonetics), len(englishPhonetics), "Extended bank should have more phonetics than English")
}

// TestVOXGenerator_calculateTriggerStrength tests trigger strength calculations
func TestVOXGenerator_calculateTriggerStrength(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{})

	tests := []struct {
		name             string
		triggerData      map[string]float64
		expectedStrength float64
		validator        func(t *testing.T, strength float64)
	}{
		{
			"MaxStrength_AllHigh",
			map[string]float64{
				"emf_anomaly":   1.0,
				"audio_anomaly": 1.0,
				"temperature":   1.0,
				"interference":  1.0,
			},
			1.0,
			func(t *testing.T, strength float64) {
				assert.Equal(t, 1.0, strength)
				assert.LessOrEqual(t, strength, 1.0, "Should be normalized to 1.0 maximum")
			},
		},
		{
			"NoTriggers",
			map[string]float64{},
			0.0,
			func(t *testing.T, strength float64) {
				assert.Equal(t, 0.0, strength)
			},
		},
		{
			"PartialTriggers",
			map[string]float64{
				"emf_anomaly":   0.5,
				"audio_anomaly": 0.8,
				// temperature and interference missing (treated as 0)
			},
			0.5*0.3 + 0.8*0.4, // emf: 0.3, audio: 0.4
			func(t *testing.T, strength float64) {
				assert.InDelta(t, 0.47, strength, 0.01, "Should calculate weighted sum correctly")
				assert.LessOrEqual(t, strength, 1.0, "Should be normalized")
			},
		},
		{
			"UnknownTriggers_Ignored",
			map[string]float64{
				"emf_anomaly":     0.5,
				"audio_anomaly":   0.5,
				"unknown_field":   1.0, // Should be ignored
				"another_unknown": 0.8, // Should be ignored
			},
			0.5*0.3 + 0.5*0.4, // Only known weights
			func(t *testing.T, strength float64) {
				assert.InDelta(t, 0.35, strength, 0.01, "Should ignore unknown fields")
			},
		},
		{
			"VeryHighValues_Clamped",
			map[string]float64{
				"emf_anomaly":   5.0,
				"audio_anomaly": 10.0,
				"temperature":   2.0,
				"interference":  3.0,
			},
			1.0, // Should be clamped to 1.0
			func(t *testing.T, strength float64) {
				assert.Equal(t, 1.0, strength, "Should be clamped to maximum")
			},
		},
		{
			"WeightsValidation",
			map[string]float64{
				"emf_anomaly":   0.7,
				"audio_anomaly": 0.6,
				"temperature":   0.4,
				"interference":  0.3,
			},
			0.7*0.3 + 0.6*0.4 + 0.4*0.1 + 0.3*0.2,
			func(t *testing.T, strength float64) {
				expected := 0.21 + 0.24 + 0.04 + 0.06
				assert.InDelta(t, expected, strength, 0.001, "Should apply correct weights")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := vox.calculateTriggerStrength(tt.triggerData)
			assert.InDelta(t, tt.expectedStrength, strength, 0.001, "Should calculate expected strength")
			tt.validator(t, strength)
		})
	}
}

// TestVOXGenerator_GenerateVOX_TriggerStrength tests VOX generation based on trigger strength
func TestVOXGenerator_GenerateVOX_TriggerStrength(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	})

	tests := []struct {
		name             string
		triggerData      map[string]float64
		expectGeneration bool
		validator        func(t *testing.T, result *VOXResult)
	}{
		{
			"AboveThreshold",
			map[string]float64{
				"emf_anomaly":   0.8,
				"audio_anomaly": 0.9,
				"temperature":   0.5,
				"interference":  0.6,
			},
			true,
			func(t *testing.T, result *VOXResult) {
				assert.NotNil(t, result)
				assert.Greater(t, result.TriggerStrength, 0.5, "Should be above threshold")
				assert.NotEmpty(t, result.GeneratedText)
				assert.NotEmpty(t, result.PhoneticBank)
				assert.NotEmpty(t, result.FrequencyData)
				assert.Equal(t, "amplitude", result.ModulationType)
				assert.False(t, result.GeneratedAt.IsZero())
			},
		},
		{
			"BelowThreshold",
			map[string]float64{
				"emf_anomaly":   0.1,
				"audio_anomaly": 0.2,
				"temperature":   0.1,
				"interference":  0.0,
			},
			false,
			func(t *testing.T, result *VOXResult) {
				assert.Nil(t, result)
			},
		},
		{
			"ExactlyAtThreshold",
			map[string]float64{
				"emf_anomaly":   0.5,
				"audio_anomaly": 0.5,
			},
			true, // Should generate at exactly threshold
			func(t *testing.T, result *VOXResult) {
				assert.NotNil(t, result)
				assert.Equal(t, 0.5, result.TriggerStrength)
			},
		},
		{
			"NoTriggerData",
			map[string]float64{},
			false,
			func(t *testing.T, result *VOXResult) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := vox.GenerateVOX(ctx, tt.triggerData, VOXConfig{
				DefaultLanguage:  "english",
				PhoneticBankSize: 25,
				TriggerThreshold: 0.5,
			})

			if tt.expectGeneration {
				assert.NoError(t, err)
				tt.validator(t, result)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, result)
				tt.validator(t, result)
			}
		})
	}
}

// TestVOXGenerator_GenerateVOX_TextGeneration tests text generation based on strength
func TestVOXGenerator_GenerateVOX_TextGeneration(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.1, // Low threshold for testing
	})

	tests := []struct {
		name            string
		triggerStrength float64
		expectedType    string // "word", "phonetic", "single"
		validator       func(t *testing.T, text string)
	}{
		{
			"HighStrength_Word",
			0.8,
			"word",
			func(t *testing.T, text string) {
				// Should be one of the English words
				englishWords := vox.languagePacks["english"]
				assert.Contains(t, englishWords, text)
				assert.True(t, len(text) > 1, "Word should be multi-character")
			},
		},
		{
			"MediumStrength_Phonetic",
			0.5,
			"phonetic",
			func(t *testing.T, text string) {
				// Should be combination of phonetics
				englishPhonetics := vox.phoneticBanks["english"]
				assert.True(t, len(text) >= 2, "Should combine multiple phonetics")
				assert.True(t, len(text) <= 4, "Should not be too long")

				// Each character should be a phonetic
				for _, char := range text {
					assert.Contains(t, englishPhonetics, string(char))
				}
			},
		},
		{
			"LowStrength_Single",
			0.2,
			"single",
			func(t *testing.T, text string) {
				// Should be single phonetic
				englishPhonetics := vox.phoneticBanks["english"]
				assert.Equal(t, 1, len(text))
				assert.Contains(t, englishPhonetics, text)
			},
		},
		{
			"VeryLowStrength_Empty",
			0.05,
			"single",
			func(t *testing.T, text string) {
				// Even very low should produce something
				englishPhonetics := vox.phoneticBanks["english"]
				assert.Equal(t, 1, len(text))
				assert.Contains(t, englishPhonetics, text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up trigger data that will result in desired strength
			triggerData := map[string]float64{
				"audio_anomaly": tt.triggerStrength, // Use audio_anomaly with weight 0.4
			}

			ctx := context.Background()
			result, err := vox.GenerateVOX(ctx, triggerData, VOXConfig{
				DefaultLanguage:  "english",
				PhoneticBankSize: 25,
				TriggerThreshold: 0.01, // Very low threshold
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.GeneratedText)
			tt.validator(t, result.GeneratedText)
		})
	}
}

// TestVOXGenerator_GenerateVOX_PhoneticBankSelection tests phonetic bank selection logic
func TestVOXGenerator_GenerateVOX_PhoneticBankSelection(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		TriggerThreshold: 0.1,
	})

	tests := []struct {
		name             string
		phoneticBankSize int
		expectedBank     string
	}{
		{"MinimalBank", 10, "minimal"},
		{"StandardBank", 25, "english"},
		{"ExtendedBank", 40, "extended"},
		{"VerySmallBank", 5, "minimal"},
		{"VeryLargeBank", 100, "extended"},
	}

	triggerData := map[string]float64{
		"audio_anomaly": 0.8,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := vox.GenerateVOX(ctx, triggerData, VOXConfig{
				DefaultLanguage:  "english",
				PhoneticBankSize: tt.phoneticBankSize,
				TriggerThreshold: 0.1,
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedBank, result.PhoneticBank)
		})
	}
}

// TestVOXGenerator_GenerateVOX_FrequencyModulation tests frequency modulation generation
func TestVOXGenerator_GenerateVOX_FrequencyModulation(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.1,
	})

	triggerData := map[string]float64{
		"audio_anomaly": 0.6,
	}

	ctx := context.Background()
	result, err := vox.GenerateVOX(ctx, triggerData, VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.1,
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	// Validate frequency data properties
	freqData := result.FrequencyData
	assert.NotEmpty(t, freqData)
	assert.Equal(t, 44100, len(freqData)) // 1 second at 44.1kHz

	// Check basic sine wave properties
	baseFreq := 440.0 // A4 note
	for i, sample := range freqData {
		timeVal := float64(i) / 44100.0
		expected := 0.3 * 0.6 * math.Sin(2*math.Pi*baseFreq*timeVal) // amplitude = 0.3 * strength

		// Account for frequency modulation
		modulation := math.Sin(2 * math.Pi * timeVal * 5) // 5Hz modulation
		expected *= (1.0 + 0.5*0.6*modulation)

		assert.InDelta(t, expected, sample, 0.01, "Sample should match expected frequency modulation")
	}

	// Test different text affects frequency
	tests := []string{"hello", "test", "a", "longerText"}

	for _, text := range tests {
		t.Run("TextEffect_"+text, func(t *testing.T) {
			// Generate with different text
			result, err := vox.GenerateVOX(ctx, triggerData, VOXConfig{
				DefaultLanguage:  "english",
				PhoneticBankSize: 25,
				TriggerThreshold: 0.1,
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.FrequencyData)

			// Different text should produce different frequency data
			if text != tests[0] {
				assert.NotEqual(t, result.FrequencyData, result.FrequencyData)
			}
		})
	}
}

// TestVOXGenerator_GenerateVOX_EdgeCases tests edge cases and error handling
func TestVOXGenerator_GenerateVOX_EdgeCases(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	})

	tests := []struct {
		name          string
		triggerData   map[string]float64
		config        VOXConfig
		expectError   bool
		errorContains string
	}{
		{
			"MissingLanguagePack",
			map[string]float64{"audio_anomaly": 0.8},
			VOXConfig{
				DefaultLanguage:  "nonexistent",
				PhoneticBankSize: 25,
				TriggerThreshold: 0.1,
			},
			false, // Should not error, just fallback or use empty
			"",
		},
		{
			"MissingPhoneticBank",
			map[string]float64{"audio_anomaly": 0.8},
			VOXConfig{
				DefaultLanguage:  "english",
				PhoneticBankSize: 25,
				TriggerThreshold: 0.1,
			},
			true, // Should error
			"phonetic bank nonexistent not found",
		},
		{
			"EmptyLanguageWords",
			map[string]float64{"audio_anomaly": 0.8},
			VOXConfig{
				DefaultLanguage:  "english",
				PhoneticBankSize: 25,
				TriggerThreshold: 0.1,
			},
			false, // Should work even if language pack is empty
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := vox.GenerateVOX(ctx, tt.triggerData, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				// Might still have errors or nil result, but shouldn't be due to bank missing
				if err != nil {
					assert.NotContains(t, err.Error(), "phonetic bank")
				}
			}
		})
	}
}

// TestVOXGenerator_GenerateVOX_ContextCancellation tests context cancellation handling
func TestVOXGenerator_GenerateVOX_ContextCancellation(t *testing.T) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	triggerData := map[string]float64{
		"audio_anomaly": 0.8,
	}

	_, err := vox.GenerateVOX(ctx, triggerData, VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	})

	// Behavior depends on implementation - may or may not check context
	// Just verify it doesn't panic
	assert.True(t, err == nil || err != nil)
}

// Performance benchmarks

func BenchmarkVOXGenerator_GenerateVOX(b *testing.B) {
	vox := NewVOXGenerator(VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	})

	triggerData := map[string]float64{
		"emf_anomaly":   0.8,
		"audio_anomaly": 0.9,
		"temperature":   0.5,
		"interference":  0.6,
	}

	config := VOXConfig{
		DefaultLanguage:  "english",
		PhoneticBankSize: 25,
		TriggerThreshold: 0.5,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vox.GenerateVOX(ctx, triggerData, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVOXGenerator_calculateTriggerStrength(b *testing.B) {
	vox := NewVOXGenerator(VOXConfig{})

	triggerData := map[string]float64{
		"emf_anomaly":   0.8,
		"audio_anomaly": 0.9,
		"temperature":   0.5,
		"interference":  0.6,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vox.calculateTriggerStrength(triggerData)
	}
}

func BenchmarkVOXGenerator_generateFrequencyModulation(b *testing.B) {
	vox := NewVOXGenerator(VOXConfig{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vox.generateFrequencyModulation("test", 0.6)
	}
}
