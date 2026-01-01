package audio

import (
	"context"
	"fmt"
	"math"
	"math/cmplx"
	"time"

	"gonum.org/v1/gonum/dsp/fourier"
)

// Processor handles audio processing for paranormal investigation
type Processor struct {
	sampleRate     int
	bitDepth       int
	noiseThreshold float64
	fft            *fourier.FFT
}

// ProcessorConfig holds configuration for audio processing
type ProcessorConfig struct {
	SampleRate     int
	BitDepth       int
	NoiseThreshold float64
}

// ProcessingResult contains the results of audio analysis
type ProcessingResult struct {
	WaveformData     []float64          `json:"waveform_data"`
	FrequencyData    []complex128       `json:"frequency_data"`
	EVPEvents        []EVPEvent         `json:"evp_events"`
	AnomalyStrength  float64            `json:"anomaly_strength"`
	NoiseLevel       float64            `json:"noise_level"`
	ProcessingTime   time.Duration      `json:"processing_time"`
	SpectralAnalysis SpectralAnalysis   `json:"spectral_analysis"`
	Metadata         ProcessingMetadata `json:"metadata"`
}

// EVPEvent represents a detected EVP event
type EVPEvent struct {
	StartTime   float64 `json:"start_time"`
	EndTime     float64 `json:"end_time"`
	Confidence  float64 `json:"confidence"`
	Frequency   float64 `json:"frequency"`
	Amplitude   float64 `json:"amplitude"`
	Description string  `json:"description"`
}

// SpectralAnalysis contains frequency domain analysis
type SpectralAnalysis struct {
	DominantFrequencies []FrequencyPeak `json:"dominant_frequencies"`
	SpectralCentroid    float64         `json:"spectral_centroid"`
	SpectralRolloff     float64         `json:"spectral_rolloff"`
	ZeroCrossingRate    float64         `json:"zero_crossing_rate"`
	RMSEnergy           float64         `json:"rms_energy"`
}

// FrequencyPeak represents a peak in the frequency spectrum
type FrequencyPeak struct {
	Frequency float64 `json:"frequency"`
	Magnitude float64 `json:"magnitude"`
	Quality   float64 `json:"quality"`
}

// ProcessingMetadata contains metadata about the processing
type ProcessingMetadata struct {
	SampleRate     int            `json:"sample_rate"`
	BitDepth       int            `json:"bit_depth"`
	Duration       float64        `json:"duration"`
	ProcessedAt    time.Time      `json:"processed_at"`
	FilterSettings FilterSettings `json:"filter_settings"`
}

// FilterSettings represents applied audio filters
type FilterSettings struct {
	HighPassCutoff float64   `json:"high_pass_cutoff"`
	LowPassCutoff  float64   `json:"low_pass_cutoff"`
	NotchFilters   []float64 `json:"notch_filters"`
	NoiseReduction bool      `json:"noise_reduction"`
	DynamicRange   bool      `json:"dynamic_range"`
}

// NewProcessor creates a new audio processor
func NewProcessor(config ProcessorConfig) *Processor {
	return &Processor{
		sampleRate:     config.SampleRate,
		bitDepth:       config.BitDepth,
		noiseThreshold: config.NoiseThreshold,
		fft:            fourier.NewFFT(1024), // Default FFT size
	}
}

// ProcessAudio processes raw audio data for paranormal analysis
func (p *Processor) ProcessAudio(ctx context.Context, audioData []float64) (*ProcessingResult, error) {
	startTime := time.Now()

	// Validate input
	if len(audioData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}

	result := &ProcessingResult{
		WaveformData: audioData,
		Metadata: ProcessingMetadata{
			SampleRate:  p.sampleRate,
			BitDepth:    p.bitDepth,
			Duration:    float64(len(audioData)) / float64(p.sampleRate),
			ProcessedAt: time.Now(),
		},
	}

	// Apply noise reduction if enabled
	filteredData := p.applyNoiseReduction(audioData)

	// Calculate basic audio metrics
	result.NoiseLevel = p.calculateNoiseLevel(filteredData)
	result.SpectralAnalysis = p.performSpectralAnalysis(filteredData)

	// Perform FFT analysis
	fftResult, err := p.performFFTAnalysis(filteredData)
	if err != nil {
		return nil, fmt.Errorf("FFT analysis failed: %w", err)
	}
	result.FrequencyData = fftResult

	// Detect EVP events
	result.EVPEvents = p.detectEVPEvents(filteredData, fftResult)

	// Calculate anomaly strength
	result.AnomalyStrength = p.calculateAnomalyStrength(filteredData, fftResult)

	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// applyNoiseReduction applies noise reduction filters to audio data
func (p *Processor) applyNoiseReduction(data []float64) []float64 {
	filtered := make([]float64, len(data))
	copy(filtered, data)

	// Apply high-pass filter to remove low-frequency noise
	filtered = p.highPassFilter(filtered, 80.0) // Remove below 80 Hz

	// Apply notch filters for common electronic interference
	commonInterference := []float64{50.0, 60.0, 120.0, 240.0} // Power line frequencies
	for _, freq := range commonInterference {
		filtered = p.notchFilter(filtered, freq, 2.0) // 2 Hz bandwidth
	}

	return filtered
}

// highPassFilter applies a simple high-pass filter
func (p *Processor) highPassFilter(data []float64, cutoffFreq float64) []float64 {
	// Simple first-order high-pass filter
	rc := 1.0 / (2.0 * math.Pi * cutoffFreq)
	dt := 1.0 / float64(p.sampleRate)
	alpha := rc / (rc + dt)

	filtered := make([]float64, len(data))
	if len(data) > 0 {
		filtered[0] = data[0]
	}

	for i := 1; i < len(data); i++ {
		filtered[i] = alpha * (filtered[i-1] + data[i] - data[i-1])
	}

	return filtered
}

// notchFilter applies a notch filter to remove specific frequency
func (p *Processor) notchFilter(data []float64, centerFreq, bandwidth float64) []float64 {
	if len(data) == 0 {
		return data
	}

	// Design a second-order IIR notch filter
	omega0 := 2.0 * math.Pi * centerFreq / float64(p.sampleRate)
	deltaOmega := 2.0 * math.Pi * bandwidth / float64(p.sampleRate)

	// Calculate filter coefficients
	cosOmega0 := math.Cos(omega0)
	alpha := math.Sin(deltaOmega/2) / 2

	b0 := 1.0
	b1 := -2.0 * cosOmega0
	b2 := 1.0
	a0 := 1.0 + alpha
	a1 := -2.0 * cosOmega0
	a2 := 1.0 - alpha

	// Normalize coefficients
	b0 /= a0
	b1 /= a0
	b2 /= a0
	a1 /= a0
	a2 /= a0

	filtered := make([]float64, len(data))

	// Apply the IIR filter with proper initialization
	if len(data) >= 1 {
		filtered[0] = b0 * data[0]
	}
	if len(data) >= 2 {
		filtered[1] = b0*data[1] + b1*data[0] - a1*filtered[0]
	}

	for i := 2; i < len(data); i++ {
		filtered[i] = b0*data[i] + b1*data[i-1] + b2*data[i-2] -
			a1*filtered[i-1] - a2*filtered[i-2]
	}

	return filtered
}

// calculateNoiseLevel calculates the overall noise level
func (p *Processor) calculateNoiseLevel(data []float64) float64 {
	if len(data) == 0 {
		return 0.0
	}

	var sum float64
	for _, sample := range data {
		sum += sample * sample
	}

	return math.Sqrt(sum / float64(len(data)))
}

// performSpectralAnalysis performs detailed spectral analysis
func (p *Processor) performSpectralAnalysis(data []float64) SpectralAnalysis {
	analysis := SpectralAnalysis{
		DominantFrequencies: []FrequencyPeak{},
	}

	if len(data) == 0 {
		return analysis
	}

	// Calculate RMS energy
	var rmsSum float64
	for _, sample := range data {
		rmsSum += sample * sample
	}
	analysis.RMSEnergy = math.Sqrt(rmsSum / float64(len(data)))

	// Calculate zero crossing rate
	crossings := 0
	for i := 1; i < len(data); i++ {
		if (data[i] >= 0) != (data[i-1] >= 0) {
			crossings++
		}
	}
	analysis.ZeroCrossingRate = float64(crossings) / float64(len(data))

	// Perform FFT for spectral features
	fftData := p.fft.Coefficients(nil, data)

	if len(fftData) > 0 {
		// Calculate spectral centroid and rolloff
		var weightedSum, magnitudeSum float64
		var magnitudes []float64

		for i := 1; i < len(fftData)/2; i++ {
			magnitude := cmplx.Abs(fftData[i])
			magnitudes = append(magnitudes, magnitude)
			freq := float64(i) * float64(p.sampleRate) / float64(len(fftData))
			weightedSum += magnitude * freq
			magnitudeSum += magnitude
		}

		if magnitudeSum > 0 {
			analysis.SpectralCentroid = weightedSum / magnitudeSum

			// Calculate spectral rolloff (95% of spectral energy)
			var cumulativeEnergy float64
			totalEnergy := magnitudeSum
			rolloffThreshold := 0.95 * totalEnergy

			for i := 1; i < len(magnitudes); i++ {
				cumulativeEnergy += magnitudes[i]
				if cumulativeEnergy >= rolloffThreshold {
					analysis.SpectralRolloff = float64(i) * float64(p.sampleRate) / float64(len(fftData))
					break
				}
			}
		}

		// Find dominant frequency peaks
		analysis.DominantFrequencies = p.findFrequencyPeaks(magnitudes, len(fftData))
	}

	return analysis
}

// findFrequencyPeaks identifies significant frequency peaks in the spectrum
func (p *Processor) findFrequencyPeaks(magnitudes []float64, fftSize int) []FrequencyPeak {
	peaks := []FrequencyPeak{}

	if len(magnitudes) < 3 {
		return peaks
	}

	// Simple peak detection algorithm
	for i := 1; i < len(magnitudes)-1; i++ {
		if magnitudes[i] > magnitudes[i-1] && magnitudes[i] > magnitudes[i+1] {
			// Peak detected
			if magnitudes[i] > p.noiseThreshold*5 { // Threshold for significant peaks
				frequency := float64(i) * float64(p.sampleRate) / float64(fftSize)
				quality := magnitudes[i] / (magnitudes[i-1] + magnitudes[i+1] + 1e-10)

				peak := FrequencyPeak{
					Frequency: frequency,
					Magnitude: magnitudes[i],
					Quality:   math.Min(quality, 1.0),
				}
				peaks = append(peaks, peak)
			}
		}
	}

	// Sort peaks by magnitude and return top 5
	for i := 0; i < len(peaks)-1; i++ {
		for j := i + 1; j < len(peaks); j++ {
			if peaks[j].Magnitude > peaks[i].Magnitude {
				peaks[i], peaks[j] = peaks[j], peaks[i]
			}
		}
	}

	if len(peaks) > 5 {
		peaks = peaks[:5]
	}

	return peaks
}

// performFFTAnalysis performs Fast Fourier Transform analysis
func (p *Processor) performFFTAnalysis(data []float64) ([]complex128, error) {
	// Ensure data length is power of 2 for efficient FFT
	fftSize := 1024
	if len(data) < fftSize {
		// Pad with zeros
		padded := make([]float64, fftSize)
		copy(padded, data)
		data = padded
	} else if len(data) > fftSize {
		// Truncate to fftSize if longer
		data = data[:fftSize]
	}

	// Use gonum's optimized FFT implementation
	fft := fourier.NewFFT(fftSize)

	// Convert to []float64 and perform FFT
	coeffs := fft.Coefficients(nil, data[:fftSize])

	return coeffs, nil
}

// detectEVPEvents detects potential EVP events in the audio
func (p *Processor) detectEVPEvents(timeData []float64, freqData []complex128) []EVPEvent {
	events := []EVPEvent{}

	if len(timeData) == 0 || len(freqData) == 0 {
		return events
	}

	// Window size for analysis (typically 50ms windows)
	windowSize := int(0.05 * float64(p.sampleRate))
	if windowSize < 256 {
		windowSize = 256
	}

	// Overlap windows by 50% for better temporal resolution
	hopSize := windowSize / 2

	// Analyze audio in sliding windows
	for start := 0; start < len(timeData)-windowSize; start += hopSize {
		end := start + windowSize
		if end > len(timeData) {
			end = len(timeData)
		}

		windowData := timeData[start:end]

		// Apply Hanning window to reduce spectral leakage
		hanningWindow := make([]float64, len(windowData))
		for i := 0; i < len(windowData); i++ {
			hanningWindow[i] = 0.5 * (1.0 - math.Cos(2.0*math.Pi*float64(i)/float64(len(windowData)-1)))
			windowData[i] *= hanningWindow[i]
		}

		// Perform FFT on windowed data
		windowFFT := p.fft.Coefficients(nil, windowData)

		// Analyze frequency spectrum for anomalies in this window
		for i := 1; i < len(windowFFT)/2; i++ {
			magnitude := cmplx.Abs(windowFFT[i])
			frequency := float64(i) * float64(p.sampleRate) / float64(len(windowFFT))

			// Look for peaks in voice frequency range
			if frequency >= 85 && frequency <= 2000 && magnitude > p.noiseThreshold*15 {
				// Calculate confidence based on magnitude and frequency characteristics
				confidence := math.Min(magnitude/(p.noiseThreshold*25), 1.0)

				if confidence > 0.4 { // Slightly higher threshold for windowed analysis
					startTime := float64(start) / float64(p.sampleRate)
					endTime := float64(end) / float64(p.sampleRate)

					event := EVPEvent{
						StartTime:   startTime,
						EndTime:     endTime,
						Confidence:  confidence,
						Frequency:   frequency,
						Amplitude:   magnitude,
						Description: fmt.Sprintf("EVP detected at %.1f Hz (%.2fs-%.2fs)", frequency, startTime, endTime),
					}
					events = append(events, event)
				}
			}
		}
	}

	// Merge overlapping events with similar frequencies
	events = p.mergeSimilarEvents(events)

	return events
}

// mergeSimilarEvents merges overlapping EVP events with similar frequencies
func (p *Processor) mergeSimilarEvents(events []EVPEvent) []EVPEvent {
	if len(events) <= 1 {
		return events
	}

	// Sort events by start time
	sortedEvents := make([]EVPEvent, len(events))
	copy(sortedEvents, events)

	// Simple merge algorithm for overlapping events with similar frequencies
	merged := []EVPEvent{}
	current := sortedEvents[0]

	for i := 1; i < len(sortedEvents); i++ {
		next := sortedEvents[i]

		// Check if events overlap and have similar frequencies
		if next.StartTime <= current.EndTime &&
			math.Abs(next.Frequency-current.Frequency) < 100.0 {
			// Merge events
			current.EndTime = math.Max(current.EndTime, next.EndTime)
			current.Confidence = math.Max(current.Confidence, next.Confidence)
			current.Amplitude = math.Max(current.Amplitude, next.Amplitude)
			current.Description = fmt.Sprintf("Merged EVP: %.1f Hz (%.2fs-%.2f)",
				current.Frequency, current.StartTime, current.EndTime)
		} else {
			merged = append(merged, current)
			current = next
		}
	}
	merged = append(merged, current)

	return merged
}

// calculateAnomalyStrength calculates the overall anomaly strength
func (p *Processor) calculateAnomalyStrength(timeData []float64, freqData []complex128) float64 {
	if len(freqData) == 0 {
		return 0.0
	}

	var totalMagnitude float64
	var voiceRangeMagnitude float64

	for i := 1; i < len(freqData)/2; i++ {
		magnitude := cmplx.Abs(freqData[i])
		frequency := float64(i) * float64(p.sampleRate) / float64(len(freqData))

		totalMagnitude += magnitude

		// Focus on human voice frequency range
		if frequency >= 85 && frequency <= 2000 {
			voiceRangeMagnitude += magnitude
		}
	}

	if totalMagnitude == 0 {
		return 0.0
	}

	// Return normalized anomaly strength
	return math.Min(voiceRangeMagnitude/totalMagnitude, 1.0)
}

// VOXGenerator handles Voice Synthesis for paranormal communication
type VOXGenerator struct {
	phoneticBanks map[string][]string
	languagePacks map[string][]string
}

// VOXConfig holds configuration for VOX generation
type VOXConfig struct {
	DefaultLanguage  string
	PhoneticBankSize int
	TriggerThreshold float64
}

// VOXResult contains the result of VOX generation
type VOXResult struct {
	GeneratedText   string    `json:"generated_text"`
	PhoneticBank    string    `json:"phonetic_bank"`
	TriggerStrength float64   `json:"trigger_strength"`
	FrequencyData   []float64 `json:"frequency_data"`
	ModulationType  string    `json:"modulation_type"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// NewVOXGenerator creates a new VOX generator
func NewVOXGenerator(config VOXConfig) *VOXGenerator {
	vox := &VOXGenerator{
		phoneticBanks: make(map[string][]string),
		languagePacks: make(map[string][]string),
	}

	// Initialize default phonetic banks
	vox.initializePhoneticBanks()
	vox.initializeLanguagePacks()

	return vox
}

// initializePhoneticBanks initializes the phonetic sound banks
func (v *VOXGenerator) initializePhoneticBanks() {
	v.phoneticBanks["english"] = []string{
		"ah", "eh", "ih", "oh", "uh", "ay", "ey", "iy", "ow", "uw",
		"b", "d", "f", "g", "h", "k", "l", "m", "n", "p", "r", "s", "t", "v", "w", "y", "z",
		"ch", "sh", "th", "ng", "zh",
	}

	v.phoneticBanks["minimal"] = []string{
		"a", "e", "i", "o", "u", "m", "n", "s", "t", "r", "l",
	}

	v.phoneticBanks["extended"] = append(v.phoneticBanks["english"],
		"aa", "ae", "ao", "aw", "ax", "er", "ia", "ua", "ai", "ei",
	)
}

// initializeLanguagePacks initializes language-specific word banks
func (v *VOXGenerator) initializeLanguagePacks() {
	v.languagePacks["english"] = []string{
		"yes", "no", "here", "there", "go", "stay", "help", "stop", "come", "leave",
		"light", "dark", "cold", "warm", "see", "hear", "feel", "know", "remember",
		"hello", "goodbye", "please", "sorry", "thank", "name", "who", "what", "when", "where",
	}

	v.languagePacks["simple"] = []string{
		"yes", "no", "go", "stop", "here", "help", "see", "hear",
	}
}

// GenerateVOX generates VOX communication based on environmental triggers
func (v *VOXGenerator) GenerateVOX(ctx context.Context, triggerData map[string]float64, config VOXConfig) (*VOXResult, error) {
	// Calculate trigger strength from environmental data
	triggerStrength := v.calculateTriggerStrength(triggerData)

	if triggerStrength < config.TriggerThreshold {
		return nil, nil // No generation below threshold
	}

	// Select phonetic bank based on configuration
	bankName := "english"
	if config.PhoneticBankSize < 20 {
		bankName = "minimal"
	} else if config.PhoneticBankSize > 30 {
		bankName = "extended"
	}

	phonetics := v.phoneticBanks[bankName]
	if len(phonetics) == 0 {
		return nil, fmt.Errorf("phonetic bank %s not found", bankName)
	}

	// Generate text based on trigger strength and randomness
	generatedText := v.generateText(phonetics, v.languagePacks[config.DefaultLanguage], triggerStrength)

	// Generate frequency modulation data
	freqData := v.generateFrequencyModulation(generatedText, triggerStrength)

	return &VOXResult{
		GeneratedText:   generatedText,
		PhoneticBank:    bankName,
		TriggerStrength: triggerStrength,
		FrequencyData:   freqData,
		ModulationType:  "amplitude",
		GeneratedAt:     time.Now(),
	}, nil
}

// calculateTriggerStrength calculates trigger strength from environmental data
func (v *VOXGenerator) calculateTriggerStrength(data map[string]float64) float64 {
	weights := map[string]float64{
		"emf_anomaly":   0.3,
		"audio_anomaly": 0.4,
		"temperature":   0.1,
		"interference":  0.2,
	}

	var totalStrength float64
	for key, value := range data {
		if weight, exists := weights[key]; exists {
			totalStrength += value * weight
		}
	}

	return math.Min(totalStrength, 1.0)
}

// generateText generates text based on phonetics and language pack
func (v *VOXGenerator) generateText(phonetics, words []string, strength float64) string {
	if strength > 0.7 && len(words) > 0 {
		// High strength: use actual words
		return words[int(strength*float64(len(words)))%len(words)]
	} else if strength > 0.4 && len(phonetics) > 0 {
		// Medium strength: combine phonetics
		count := int(strength*3) + 1
		result := ""
		for i := 0; i < count; i++ {
			if i > 0 {
				result += ""
			}
			result += phonetics[int(strength*float64(len(phonetics))*float64(i+1))%len(phonetics)]
		}
		return result
	}

	// Low strength: single phonetic
	if len(phonetics) > 0 {
		return phonetics[int(strength*float64(len(phonetics)))%len(phonetics)]
	}

	return ""
}

// generateFrequencyModulation generates frequency data for VOX output
func (v *VOXGenerator) generateFrequencyModulation(text string, strength float64) []float64 {
	baseFreq := 440.0 // A4 note
	duration := 1.0   // 1 second
	sampleRate := 44100.0

	samples := int(duration * sampleRate)
	data := make([]float64, samples)

	for i := 0; i < samples; i++ {
		t := float64(i) / sampleRate

		// Modulate frequency based on text and strength
		freq := baseFreq * (1.0 + strength*0.5*math.Sin(2*math.Pi*t*5))
		amplitude := 0.3 * strength

		data[i] = amplitude * math.Sin(2*math.Pi*freq*t)
	}

	return data
}
