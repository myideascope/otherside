package audio

import (
	"context"
	"fmt"
	"math"
	"math/cmplx"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test audio data fixtures
var (
	// Silent audio (all zeros)
	silentAudio = make([]float64, 1000)

	// Pure sine wave at 440Hz for 1 second at 44100Hz sample rate
	sineWave440 = func() []float64 {
		const sampleRate = 44100
		const duration = 1.0
		const frequency = 440.0

		samples := int(sampleRate * duration)
		data := make([]float64, samples)
		for i := 0; i < samples; i++ {
			t := float64(i) / sampleRate
			data[i] = math.Sin(2 * math.Pi * frequency * t)
		}
		return data
	}()

	// Voice frequency range simulation (85-2000 Hz)
	voiceRange = func() []float64 {
		const sampleRate = 44100
		const duration = 1.0

		samples := int(sampleRate * duration)
		data := make([]float64, samples)
		for i := 0; i < samples; i++ {
			t := float64(i) / sampleRate
			// Combine multiple frequencies in voice range
			data[i] = 0.3*math.Sin(2*math.Pi*150*t) + // Low voice
				0.2*math.Sin(2*math.Pi*440*t) + // Mid voice
				0.1*math.Sin(2*math.Pi*1200*t) // High voice
		}
		return data
	}()

	// Noise with frequency peaks
	noisyWithPeaks = func() []float64 {
		const sampleRate = 44100
		const duration = 1.0

		samples := int(sampleRate * duration)
		data := make([]float64, samples)
		for i := 0; i < samples; i++ {
			t := float64(i) / sampleRate
			// White noise with periodic peaks
			data[i] = 0.1*(2*float64(i%100)/100-1) + // Noise
				0.5*math.Sin(2*math.Pi*440*t) // Strong peak
		}
		return data
	}()

	// Short audio for edge case testing
	shortAudio = []float64{0.1, -0.2, 0.3, -0.1, 0.2}

	// Single sample
	singleSample = []float64{0.5}
)

// TestNewProcessor validates processor creation and configuration
func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name   string
		config ProcessorConfig
	}{
		{"StandardConfig", ProcessorConfig{SampleRate: 44100, BitDepth: 16, NoiseThreshold: 0.1}},
		{"HighQuality", ProcessorConfig{SampleRate: 48000, BitDepth: 24, NoiseThreshold: 0.05}},
		{"LowThreshold", ProcessorConfig{SampleRate: 22050, BitDepth: 8, NoiseThreshold: 0.01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(tt.config)

			assert.NotNil(t, processor)
			assert.Equal(t, tt.config.SampleRate, processor.sampleRate)
			assert.Equal(t, tt.config.BitDepth, processor.bitDepth)
			assert.Equal(t, tt.config.NoiseThreshold, processor.noiseThreshold)
			assert.NotNil(t, processor.fft)
		})
	}
}

// TestAudioProcessor_ProcessAudio_CoreProcessing tests core audio processing functionality
func TestAudioProcessor_ProcessAudio_CoreProcessing(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{
		SampleRate:     44100,
		BitDepth:       16,
		NoiseThreshold: 0.1,
	})

	tests := []struct {
		name          string
		audioData     []float64
		expectError   bool
		expectedError string
	}{
		{"EmptyAudio", []float64{}, true, "empty audio data"},
		{"NilAudio", nil, true, "empty audio data"},
		{"SilentAudio", func() []float64 { return make([]float64, 100) }(), false, ""},
		{"ValidAudio", sineWave440, false, ""},
		{"ShortAudio", shortAudio, false, ""},
		{"SingleSample", singleSample, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := processor.ProcessAudio(ctx, tt.audioData)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.audioData, result.WaveformData)
				assert.GreaterOrEqual(t, result.ProcessingTime, time.Duration(0))
				assert.Equal(t, processor.sampleRate, result.Metadata.SampleRate)
				assert.Equal(t, processor.bitDepth, result.Metadata.BitDepth)
				assert.InDelta(t, float64(len(tt.audioData))/float64(processor.sampleRate),
					result.Metadata.Duration, 0.001)
			}
		})
	}
}

// TestAudioProcessor_ProcessAudio_ContextCancellation tests context cancellation handling
func TestAudioProcessor_ProcessAudio_ContextCancellation(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{
		SampleRate:     44100,
		BitDepth:       16,
		NoiseThreshold: 0.1,
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	_, err := processor.ProcessAudio(ctx, sineWave440)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestAudioProcessor_ProcessAudio_InvalidSampleRates tests various sample rates
func TestAudioProcessor_ProcessAudio_InvalidSampleRates(t *testing.T) {
	invalidRates := []int{0, -1, 1, 100, 10000000}

	for _, rate := range invalidRates {
		t.Run(fmt.Sprintf("InvalidRate_%d", rate), func(t *testing.T) {
			processor := NewProcessor(ProcessorConfig{
				SampleRate:     rate,
				BitDepth:       16,
				NoiseThreshold: 0.1,
			})

			// Should still create processor but may fail during processing
			result, err := processor.ProcessAudio(context.Background(), sineWave440)

			// Either succeeds (if internally handled) or fails gracefully
			if err != nil {
				assert.NotContains(t, err.Error(), "panic")
			}
			if result != nil {
				assert.NotNil(t, result.Metadata)
			}
		})
	}
}

// TestAudioProcessor_applyNoiseReduction tests noise reduction filters
func TestAudioProcessor_applyNoiseReduction(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	tests := []struct {
		name      string
		data      []float64
		changes   bool // Should be different from input
		preserved bool // Key characteristics should be preserved
	}{
		{"SilentAudio", silentAudio, false, false},
		{"SineWave440", sineWave440, true, true}, // Filters applied but signal preserved
		{"VoiceRange", voiceRange, true, true},
		{"NoisyWithPeaks", noisyWithPeaks, true, true},
		{"ShortAudio", shortAudio, true, true},
		{"SingleSample", singleSample, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.applyNoiseReduction(tt.data)

			assert.Len(t, result, len(tt.data), "Output length should match input length")

			if tt.changes {
				// Check that some filtering occurred
				different := false
				for i := range tt.data {
					if math.Abs(result[i]-tt.data[i]) > 1e-10 {
						different = true
						break
					}
				}
				assert.True(t, different, "Filtered data should be different from input")
			}

			if tt.preserved {
				// Check that major signal characteristics are preserved
				inputRMS := calculateRMS(tt.data)
				outputRMS := calculateRMS(result)

				// RMS should be somewhat preserved (within reasonable bounds)
				assert.Greater(t, outputRMS, inputRMS*0.1)
				assert.Less(t, outputRMS, inputRMS*2.0)
			}
		})
	}
}

// TestAudioProcessor_highPassFilter tests high-pass filter functionality
func TestAudioProcessor_highPassFilter(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	tests := []struct {
		name       string
		data       []float64
		cutoffFreq float64
		validator  func(t *testing.T, input, output []float64, cutoff float64)
	}{
		{
			"LowFrequencyAttenuation",
			generateSineWave(50, 44100, 1.0), // 50Hz sine wave
			80.0,
			func(t *testing.T, input, output []float64, cutoff float64) {
				inputRMS := calculateRMS(input)
				outputRMS := calculateRMS(output)

				// Low frequency should be significantly attenuated
				assert.Less(t, outputRMS, inputRMS*0.5,
					"Low frequency should be attenuated by high-pass filter")
			},
		},
		{
			"HighFrequencyPreservation",
			generateSineWave(1000, 44100, 1.0), // 1000Hz sine wave
			80.0,
			func(t *testing.T, input, output []float64, cutoff float64) {
				inputRMS := calculateRMS(input)
				outputRMS := calculateRMS(output)

				// High frequency should be mostly preserved
				assert.Greater(t, outputRMS, inputRMS*0.7,
					"High frequency should be preserved by high-pass filter")
			},
		},
		{
			"EdgeCase_SingleSample",
			[]float64{1.0},
			80.0,
			func(t *testing.T, input, output []float64, cutoff float64) {
				assert.Len(t, output, 1)
				assert.Equal(t, input[0], output[0])
			},
		},
		{
			"EdgeCase_EmptyData",
			[]float64{},
			80.0,
			func(t *testing.T, input, output []float64, cutoff float64) {
				assert.Empty(t, output)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processor.highPassFilter(tt.data, tt.cutoffFreq)
			tt.validator(t, tt.data, result, tt.cutoffFreq)
		})
	}
}

// TestAudioProcessor_notchFilter tests notch filter functionality
func TestAudioProcessor_notchFilter(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	commonFreqs := []float64{50.0, 60.0, 120.0, 240.0} // Power line frequencies

	for _, targetFreq := range commonFreqs {
		t.Run(fmt.Sprintf("NotchAt%.0fHz", targetFreq), func(t *testing.T) {
			// Generate signal at target frequency
			signal := generateSineWave(targetFreq, 44100, 1.0)

			// Apply notch filter
			filtered := processor.notchFilter(signal, targetFreq, 2.0)

			inputRMS := calculateRMS(signal)
			outputRMS := calculateRMS(filtered)

			// Target frequency should be significantly attenuated
			assert.Less(t, outputRMS, inputRMS*0.3,
				fmt.Sprintf("Frequency at %.0f Hz should be attenuated by notch filter", targetFreq))
		})
	}

	// Test edge cases
	t.Run("EmptyData", func(t *testing.T) {
		result := processor.notchFilter([]float64{}, 60.0, 2.0)
		assert.Empty(t, result)
	})

	t.Run("SingleSample", func(t *testing.T) {
		result := processor.notchFilter([]float64{1.0}, 60.0, 2.0)
		assert.Len(t, result, 1)
	})

	t.Run("VeryShortData", func(t *testing.T) {
		data := []float64{0.1, 0.2}
		result := processor.notchFilter(data, 60.0, 2.0)
		assert.Len(t, result, 2)
	})
}

// TestAudioProcessor_notchFilter_CoefficientValidation validates IIR filter coefficients
func TestAudioProcessor_notchFilter_CoefficientValidation(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	testCases := []struct {
		freq      float64
		bandwidth float64
	}{
		{50.0, 2.0},
		{60.0, 2.0},
		{100.0, 5.0},
		{1000.0, 10.0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Coefficients_%.0fHz_%.1fBW", tc.freq, tc.bandwidth), func(t *testing.T) {
			// Generate test signal
			data := generateSineWave(tc.freq, 44100, 1.0)

			// Apply filter
			result := processor.notchFilter(data, tc.freq, tc.bandwidth)

			// Check that filter is stable (no infinite values)
			for i, sample := range result {
				assert.False(t, math.IsNaN(sample), "NaN detected at index %d", i)
				assert.False(t, math.IsInf(sample, 0), "Inf detected at index %d", i)
				assert.Less(t, math.Abs(sample), 1e6, "Unreasonably large value at index %d", i)
			}

			// Check that filter actually filters
			inputEnergy := calculateRMS(data)
			outputEnergy := calculateRMS(result)
			assert.Less(t, outputEnergy, inputEnergy*0.5, "Filter should attenuate target frequency")
		})
	}
}

// TestAudioProcessor_performFFTAnalysis tests FFT analysis functionality
func TestAudioProcessor_performFFTAnalysis(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	tests := []struct {
		name          string
		data          []float64
		expectedSize  int
		validatePeaks bool
	}{
		{"DefaultSize", sineWave440[:512], 1024, true},
		{"PowerOfTwo", sineWave440[:1024], 1024, true},
		{"NonPowerOfTwo", sineWave440[:1000], 1024, true}, // Should pad to 1024
		{"LargeData", sineWave440[:2048], 1024, true},     // Should truncate to 1024
		{"SmallData", sineWave440[:100], 1024, false},     // Should pad to 1024
		{"SingleSample", singleSample, 1024, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.performFFTAnalysis(tt.data)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedSize, len(result), "FFT size should be correct")

			// Validate FFT properties
			if tt.validatePeaks && len(tt.data) >= 256 {
				// Find frequency with maximum magnitude
				maxMag := 0.0
				maxFreqIndex := 0

				for i := 1; i < len(result)/2; i++ {
					mag := cmplx.Abs(result[i])
					if mag > maxMag {
						maxMag = mag
						maxFreqIndex = i
					}
				}

				// Convert index to frequency
				maxFreq := float64(maxFreqIndex) * float64(processor.sampleRate) / float64(len(result))

				// Should find frequency close to 440Hz for sine wave input
				if tt.name == "DefaultSize" || tt.name == "PowerOfTwo" {
					assert.InDelta(t, 440.0, maxFreq, 50.0,
						"FFT should detect frequency near 440Hz")
				}
			}
		})
	}

	// Test FFT accuracy
	t.Run("FFT_Accuracy", func(t *testing.T) {
		// Generate pure sine wave
		const testFreq = 440.0
		const testAmp = 1.0
		const sampleRate = 44100.0
		const duration = 1.0

		samples := int(sampleRate * duration)
		data := make([]float64, samples)
		for i := 0; i < samples; i++ {
			t := float64(i) / sampleRate
			data[i] = testAmp * math.Sin(2*math.Pi*testFreq*t)
		}

		result, err := processor.performFFTAnalysis(data)
		require.NoError(t, err)

		// Find peak corresponding to 440Hz
		expectedBin := int((testFreq * float64(len(result))) / float64(sampleRate))

		// Check magnitude at expected frequency bin
		expectedMag := cmplx.Abs(result[expectedBin])
		assert.Greater(t, expectedMag, 0.0, "Expected frequency should have non-zero magnitude")

		// Check nearby bins for comparison
		noiseFloor := 0.0
		for i := 1; i < len(result)/2; i++ {
			if i != expectedBin {
				mag := cmplx.Abs(result[i])
				noiseFloor += mag
			}
		}
		noiseFloor /= float64(len(result)/2 - 1)

		// Signal should be significantly above noise floor
		assert.Greater(t, expectedMag, noiseFloor*10,
			"Signal should be at least 10x above noise floor")
	})
}

// TestAudioProcessor_performSpectralAnalysis tests spectral analysis functionality
func TestAudioProcessor_performSpectralAnalysis(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	tests := []struct {
		name      string
		data      []float64
		validator func(t *testing.T, analysis SpectralAnalysis)
	}{
		{
			"SilentAudio",
			silentAudio,
			func(t *testing.T, analysis SpectralAnalysis) {
				assert.Equal(t, 0.0, analysis.RMSEnergy)
				assert.Equal(t, 0.0, analysis.ZeroCrossingRate)
				assert.Empty(t, analysis.DominantFrequencies)
			},
		},
		{
			"SineWave440",
			sineWave440[:2048],
			func(t *testing.T, analysis SpectralAnalysis) {
				assert.Greater(t, analysis.RMSEnergy, 0.0)
				assert.Greater(t, analysis.ZeroCrossingRate, 0.0)

				// Should detect dominant frequency near 440Hz
				found := false
				for _, peak := range analysis.DominantFrequencies {
					if math.Abs(peak.Frequency-440.0) < 50.0 {
						found = true
						assert.Greater(t, peak.Magnitude, 0.0)
						assert.GreaterOrEqual(t, peak.Quality, 0.0)
						assert.LessOrEqual(t, peak.Quality, 1.0)
						break
					}
				}
				assert.True(t, found, "Should detect frequency near 440Hz")
			},
		},
		{
			"VoiceRange",
			voiceRange[:2048],
			func(t *testing.T, analysis SpectralAnalysis) {
				assert.Greater(t, analysis.RMSEnergy, 0.0)
				assert.Greater(t, analysis.ZeroCrossingRate, 0.0)

				// Spectral centroid should be in voice frequency range
				assert.GreaterOrEqual(t, analysis.SpectralCentroid, 85.0)
				assert.LessOrEqual(t, analysis.SpectralCentroid, 2000.0)

				// Spectral rolloff should be reasonable
				assert.Greater(t, analysis.SpectralRolloff, 0.0)
				assert.Less(t, analysis.SpectralRolloff, float64(processor.sampleRate)/2)
			},
		},
		{
			"ShortAudio",
			shortAudio,
			func(t *testing.T, analysis SpectralAnalysis) {
				assert.Greater(t, analysis.RMSEnergy, 0.0)

				// Zero crossing rate calculation works with short data
				assert.GreaterOrEqual(t, analysis.ZeroCrossingRate, 0.0)
				assert.LessOrEqual(t, analysis.ZeroCrossingRate, 1.0)
			},
		},
		{
			"EmptyData",
			[]float64{},
			func(t *testing.T, analysis SpectralAnalysis) {
				assert.Equal(t, 0.0, analysis.RMSEnergy)
				assert.Equal(t, 0.0, analysis.ZeroCrossingRate)
				assert.Empty(t, analysis.DominantFrequencies)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := processor.performSpectralAnalysis(tt.data)
			tt.validator(t, analysis)
		})
	}
}

// TestAudioProcessor_findFrequencyPeaks tests frequency peak detection
func TestAudioProcessor_findFrequencyPeaks(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16, NoiseThreshold: 0.1})

	tests := []struct {
		name          string
		magnitudes    []float64
		fftSize       int
		expectedPeaks int
		maxPeakCount  int
		validator     func(t *testing.T, peaks []FrequencyPeak)
	}{
		{
			"EmptyMagnitudes",
			[]float64{},
			1024,
			0,
			5,
			nil,
		},
		{
			"SinglePeak",
			func() []float64 {
				mags := make([]float64, 512)
				mags[100] = 1.0 // Single peak
				return mags
			}(),
			1024,
			1,
			5,
			func(t *testing.T, peaks []FrequencyPeak) {
				assert.Len(t, peaks, 1)
				if len(peaks) == 1 {
					assert.Greater(t, peaks[0].Frequency, 0.0)
					assert.Greater(t, peaks[0].Magnitude, 0.0)
					assert.GreaterOrEqual(t, peaks[0].Quality, 0.0)
					assert.LessOrEqual(t, peaks[0].Quality, 1.0)
				}
			},
		},
		{
			"MultiplePeaks",
			func() []float64 {
				mags := make([]float64, 512)
				mags[50] = 0.8  // Peak 1
				mags[100] = 1.0 // Peak 2 (highest)
				mags[150] = 0.6 // Peak 3

				// Add noise
				for i := range mags {
					mags[i] += 0.01 * (2*float64(i%100)/100 - 1)
				}
				return mags
			}(),
			1024,
			3,
			5,
			func(t *testing.T, peaks []FrequencyPeak) {
				assert.Len(t, peaks, 3)

				// Should be sorted by magnitude (highest first)
				for i := 1; i < len(peaks); i++ {
					assert.GreaterOrEqual(t, peaks[i-1].Magnitude, peaks[i].Magnitude,
						"Peaks should be sorted by magnitude in descending order")
				}
			},
		},
		{
			"TooManyPeaks",
			func() []float64 {
				mags := make([]float64, 100)

				// Create many peaks above threshold
				for i := 10; i < 90; i += 5 {
					mags[i] = 0.6 + 0.01*float64(i)
				}
				return mags
			}(),
			512,
			5, // Should be limited to 5
			5,
			nil,
		},
		{
			"BelowThreshold",
			[]float64{0.01, 0.02, 0.01, 0.02, 0.01}, // All below threshold
			512,
			0,
			5,
			nil,
		},
		{
			"VeryShort",
			[]float64{0.1, 0.5, 0.1}, // Minimum length for peak detection
			8,
			1,
			5,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			peaks := processor.findFrequencyPeaks(tt.magnitudes, tt.fftSize)

			if tt.expectedPeaks == 0 {
				assert.Empty(t, peaks)
			} else {
				assert.LessOrEqual(t, len(peaks), tt.maxPeakCount)
				if tt.expectedPeaks <= tt.maxPeakCount {
					assert.Equal(t, tt.expectedPeaks, len(peaks))
				}
			}

			// Validate peak properties
			for _, peak := range peaks {
				assert.GreaterOrEqual(t, peak.Frequency, 0.0)
				assert.Greater(t, peak.Magnitude, 0.0)
				assert.GreaterOrEqual(t, peak.Quality, 0.0)
				assert.LessOrEqual(t, peak.Quality, 1.0)
			}

			// Check sorting by magnitude
			for i := 1; i < len(peaks); i++ {
				assert.GreaterOrEqual(t, peaks[i-1].Magnitude, peaks[i].Magnitude,
					"Peaks should be sorted by magnitude in descending order")
			}

			if tt.validator != nil {
				tt.validator(t, peaks)
			}
		})
	}
}

// TestAudioProcessor_detectEVPEvents tests EVP event detection
func TestAudioProcessor_detectEVPEvents(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16, NoiseThreshold: 0.1})

	tests := []struct {
		name           string
		timeData       []float64
		freqData       []complex128
		expectedEvents int
		validator      func(t *testing.T, events []EVPEvent)
	}{
		{
			"SilentAudio",
			silentAudio,
			make([]complex128, 1024),
			0,
			nil,
		},
		{
			"VoiceFrequencySignal",
			func() []float64 {
				// Generate signal with voice frequency components
				data := make([]float64, 2048)
				for i := range data {
					t := float64(i) / 44100
					data[i] = 0.5 * math.Sin(2*math.Pi*200*t) // Voice frequency
				}
				return data
			}(),
			func() []complex128 {
				// Create simple frequency data for testing
				data := make([]complex128, 1024)
				// Add a voice frequency component
				voiceFreq := float64(200)
				voiceBin := int(voiceFreq * 1024.0 / 44100.0)
				data[voiceBin] = complex(2.0, 0)
				return data
			}(),
			1,
			func(t *testing.T, events []EVPEvent) {
				assert.Len(t, events, 1)
				if len(events) == 1 {
					event := events[0]
					assert.GreaterOrEqual(t, event.Frequency, 85.0)
					assert.LessOrEqual(t, event.Frequency, 2000.0)
					assert.GreaterOrEqual(t, event.Confidence, 0.4)
					assert.LessOrEqual(t, event.Confidence, 1.0)
					assert.GreaterOrEqual(t, event.StartTime, 0.0)
					assert.Greater(t, event.EndTime, event.StartTime)
					assert.NotEmpty(t, event.Description)
				}
			},
		},
		{
			"EmptyData",
			[]float64{},
			[]complex128{},
			0,
			nil,
		},
		{
			"LowFrequencyOnly",
			generateSineWave(50, 44100, 1.0), // Below voice range
			make([]complex128, 1024),
			0,
			nil,
		},
		{
			"HighFrequencyOnly",
			generateSineWave(3000, 44100, 1.0), // Above voice range
			make([]complex128, 1024),
			0,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := processor.detectEVPEvents(tt.timeData, tt.freqData)
			assert.Equal(t, tt.expectedEvents, len(events))

			// Validate event properties
			for _, event := range events {
				assert.GreaterOrEqual(t, event.StartTime, 0.0)
				assert.Greater(t, event.EndTime, event.StartTime)
				assert.GreaterOrEqual(t, event.Confidence, 0.0)
				assert.LessOrEqual(t, event.Confidence, 1.0)
				assert.GreaterOrEqual(t, event.Frequency, 0.0)
				assert.Greater(t, event.Amplitude, 0.0)
				assert.NotEmpty(t, event.Description)
			}

			if tt.validator != nil {
				tt.validator(t, events)
			}
		})
	}
}

// TestAudioProcessor_mergeSimilarEvents tests event merging functionality
func TestAudioProcessor_mergeSimilarEvents(t *testing.T) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	tests := []struct {
		name           string
		events         []EVPEvent
		expectedEvents int
		validator      func(t *testing.T, merged []EVPEvent)
	}{
		{
			"NoEvents",
			[]EVPEvent{},
			0,
			nil,
		},
		{
			"SingleEvent",
			[]EVPEvent{{
				StartTime: 0.0, EndTime: 1.0, Frequency: 440.0, Confidence: 0.5,
			}},
			1,
			nil,
		},
		{
			"NonOverlappingEvents",
			[]EVPEvent{
				{StartTime: 0.0, EndTime: 1.0, Frequency: 440.0, Confidence: 0.5},
				{StartTime: 2.0, EndTime: 3.0, Frequency: 880.0, Confidence: 0.6},
			},
			2,
			nil,
		},
		{
			"OverlappingSimilarFrequency",
			[]EVPEvent{
				{StartTime: 0.0, EndTime: 1.0, Frequency: 440.0, Confidence: 0.5, Amplitude: 0.1},
				{StartTime: 0.5, EndTime: 1.5, Frequency: 450.0, Confidence: 0.6, Amplitude: 0.15, Description: "Test 2"},
			},
			1,
			func(t *testing.T, merged []EVPEvent) {
				assert.Len(t, merged, 1)
				if len(merged) == 1 {
					event := merged[0]
					assert.Equal(t, 0.0, event.StartTime)
					assert.Equal(t, 1.5, event.EndTime)    // Merged end time
					assert.Equal(t, 0.6, event.Confidence) // Max confidence
					assert.Equal(t, 0.15, event.Amplitude) // Max amplitude
					assert.Contains(t, event.Description, "Merged EVP")
				}
			},
		},
		{
			"OverlappingDifferentFrequency",
			[]EVPEvent{
				{StartTime: 0.0, EndTime: 1.0, Frequency: 440.0, Confidence: 0.5},
				{StartTime: 0.5, EndTime: 1.5, Frequency: 600.0, Confidence: 0.6}, // >100Hz difference
			},
			2,
			nil,
		},
		{
			"MultipleOverlappingEvents",
			[]EVPEvent{
				{StartTime: 0.0, EndTime: 0.5, Frequency: 440.0, Confidence: 0.4, Amplitude: 0.1},
				{StartTime: 0.3, EndTime: 0.8, Frequency: 445.0, Confidence: 0.5, Amplitude: 0.12},
				{StartTime: 0.6, EndTime: 1.0, Frequency: 450.0, Confidence: 0.6, Amplitude: 0.14},
			},
			1,
			func(t *testing.T, merged []EVPEvent) {
				assert.Len(t, merged, 1)
				if len(merged) == 1 {
					event := merged[0]
					assert.Equal(t, 0.0, event.StartTime)
					assert.Equal(t, 1.0, event.EndTime)
					assert.Equal(t, 0.6, event.Confidence) // Max confidence
					assert.Equal(t, 0.14, event.Amplitude) // Max amplitude
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := processor.mergeSimilarEvents(tt.events)
			assert.Equal(t, tt.expectedEvents, len(merged))

			// Validate merged events
			for _, event := range merged {
				assert.GreaterOrEqual(t, event.StartTime, 0.0)
				assert.Greater(t, event.EndTime, event.StartTime)
				assert.GreaterOrEqual(t, event.Confidence, 0.0)
				assert.LessOrEqual(t, event.Confidence, 1.0)
				assert.GreaterOrEqual(t, event.Frequency, 0.0)
			}

			if tt.validator != nil {
				tt.validator(t, merged)
			}
		})
	}
}

// Helper functions for testing

func generateSineWave(frequency, sampleRate, duration float64) []float64 {
	samples := int(sampleRate * duration)
	data := make([]float64, samples)
	for i := 0; i < samples; i++ {
		t := float64(i) / sampleRate
		data[i] = math.Sin(2 * math.Pi * frequency * t)
	}
	return data
}

func calculateRMS(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}
	var sum float64
	for _, sample := range data {
		sum += sample * sample
	}
	return math.Sqrt(sum / float64(len(data)))
}

// Performance benchmarks

func BenchmarkAudioProcessor_ProcessAudio(b *testing.B) {
	processor := NewProcessor(ProcessorConfig{
		SampleRate:     44100,
		BitDepth:       16,
		NoiseThreshold: 0.1,
	})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessAudio(ctx, sineWave440[:1024])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAudioProcessor_applyNoiseReduction(b *testing.B) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.applyNoiseReduction(sineWave440[:1024])
	}
}

func BenchmarkAudioProcessor_performFFTAnalysis(b *testing.B) {
	processor := NewProcessor(ProcessorConfig{SampleRate: 44100, BitDepth: 16})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.performFFTAnalysis(sineWave440[:1024])
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAudioProcessor_detectEVPEvents(b *testing.B) {
	processor := NewProcessor(ProcessorConfig{
		SampleRate:     44100,
		BitDepth:       16,
		NoiseThreshold: 0.1,
	})

	timeData := generateSineWave(200, 44100, 1.0) // Voice frequency
	freqData := make([]complex128, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.detectEVPEvents(timeData[:1024], freqData)
	}
}
