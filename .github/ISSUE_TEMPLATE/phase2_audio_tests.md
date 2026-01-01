# Phase 2: Audio Processing Tests

## Description
Implement comprehensive unit tests for audio processing algorithms including EVP detection, VOX generation, and spectral analysis. This phase validates the core paranormal detection functionality.

## Tasks

### 1. Audio Processor Tests (`pkg/audio/processor_test.go`)
- [ ] **Core Processing Tests**
  - Audio data validation (empty, nil, invalid sample rates)
  - ProcessingResult structure validation
  - Context cancellation handling
  - Processing time measurement validation

- [ ] **Noise Reduction Tests**
  - High-pass filter effectiveness at 80Hz cutoff
  - Notch filter frequency accuracy (50Hz, 60Hz, 120Hz, 240Hz)
  - Filter coefficient calculations
  - Edge case handling (single sample, short arrays)
  - Filter stability tests

- [ ] **FFT Analysis Tests**
  - FFT size handling (power of 2, padding, truncation)
  - Frequency calculation accuracy
  - Magnitude spectrum validation
  - Window function effects
  - Performance with large datasets

- [ ] **Spectral Analysis Tests**
  - Spectral centroid calculation accuracy
  - Spectral rolloff at 95% energy threshold
  - Zero crossing rate calculation
  - RMS energy computation
  - Frequency peak detection algorithm

- [ ] **Peak Detection Tests**
  - Peak identification accuracy
  - Noise threshold filtering
  - Peak ranking by magnitude
  - Quality score calculation
  - Maximum peak count (5 peaks)

- [ ] **EVP Detection Tests**
  - Voice frequency range detection (85-2000 Hz)
  - Window-based analysis (50ms windows, 50% overlap)
  - Hanning window application
  - Confidence threshold validation
  - Event merging logic for overlapping events

- [ ] **Filter Operations Tests**
  - IIR filter coefficient calculations
  - Filter response characteristics
  - Numeric stability and overflow protection
  - Boundary condition handling

### 2. VOX Generator Tests (`pkg/audio/vox_test.go`)
- [ ] **Initialization Tests**
  - Phonetic bank loading (english, minimal, extended)
  - Language pack validation
  - Default configuration setup
  - Missing bank handling

- [ ] **Trigger Strength Tests**
  - Weight calculation accuracy (emf: 0.3, audio: 0.4, temp: 0.1, interference: 0.2)
  - Environmental data validation
  - Threshold comparison logic
  - Normalization to 0.0-1.0 range

- [ ] **Text Generation Tests**
  - High strength word selection (>0.7)
  - Medium strength phonetic combination (>0.4)
  - Low strength single phonetic generation
  - Empty bank handling
  - Random seed reproducibility

- [ ] **Frequency Modulation Tests**
  - Base frequency (440Hz A4 note)
  - Amplitude modulation calculation
  - Sample rate handling (44100Hz)
  - Duration generation (1 second)
  - Text-based frequency variation

## Test Audio Data Requirements
```go
// Test audio samples
var (
    // Silent audio (all zeros)
    silentAudio = []float64{0, 0, 0, ...}
    
    // Pure sine wave at 440Hz
    sineWave440 = generateSineWave(440, 44100, 1.0)
    
    // Voice frequency range simulation
    voiceRange = generateVoiceFrequencyAudio()
    
    // Noise with frequency peaks
    noisyWithPeaks = addNoiseToAudio(sineWave440, 0.1)
)
```

## Mathematical Validation Tests
- [ ] FFT accuracy vs numpy reference
- [ ] Filter frequency response curves
- [ ] Window function mathematical properties
- [ ] Statistical analysis of noise reduction
- [ ] Frequency peak detection precision

## Performance Benchmarks
- [ ] Processing time vs audio length
- [ ] Memory usage during processing
- [ ] FFT performance with different sizes
- [ ] Real-time processing capability (>10Hz update rate)

## Acceptance Criteria
- [ ] All audio algorithms have 95%+ line coverage
- [ ] Mathematical accuracy verified against reference implementations
- [ ] Performance meets real-time requirements (<100ms for 1s audio)
- [ ] Edge cases handled gracefully
- [ ] Memory usage stable and no leaks
- [ ] EVP detection minimizes false positives

## Technical Implementation
```go
func TestAudioProcessor_ProcessAudio_EVPDetection(t *testing.T) {
    tests := []struct {
        name           string
        audioData      []float64
        expectedEvents  int
        minConfidence  float64
    }{
        {"SilentAudio", silentAudio, 0, 0.4},
        {"VoiceFrequency", voiceRange, 2, 0.4},
        {"NoiseWithPeaks", noisyWithPeaks, 1, 0.4},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            processor := NewProcessor(ProcessorConfig{
                SampleRate:     44100,
                BitDepth:       16,
                NoiseThreshold: 0.1,
            })
            
            result, err := processor.ProcessAudio(context.Background(), tt.audioData)
            require.NoError(t, err)
            assert.Len(t, result.EVPEvents, tt.expectedEvents)
            
            for _, event := range result.EVPEvents {
                assert.GreaterOrEqual(t, event.Confidence, tt.minConfidence)
                assert.GreaterOrEqual(t, event.Frequency, 85.0)
                assert.LessOrEqual(t, event.Frequency, 2000.0)
            }
        })
    }
}
```

## Labels
`testing`, `audio-processing`, `evp`, `vox`, `phase-2`
`enhancement` `testing` `audio` `algorithms` `performance` `phase-2`

## Priority
High - Core paranormal detection functionality
## Estimated Effort
5-6 days (complex mathematical algorithms)
## Dependencies
- Phase 1: Domain & Configuration Tests completed
- Test infrastructure setup completed