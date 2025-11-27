package domain

import (
	"time"
)

// Session represents a paranormal investigation session
type Session struct {
	ID            string            `json:"id" db:"id"`
	Title         string            `json:"title" db:"title"`
	Location      Location          `json:"location" db:"location"`
	StartTime     time.Time         `json:"start_time" db:"start_time"`
	EndTime       *time.Time        `json:"end_time,omitempty" db:"end_time"`
	Notes         string            `json:"notes" db:"notes"`
	Environmental Environmental     `json:"environmental" db:"environmental"`
	Status        SessionStatus     `json:"status" db:"status"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
	EVPRecordings []EVPRecording    `json:"evp_recordings,omitempty"`
	VOXEvents     []VOXEvent        `json:"vox_events,omitempty"`
	RadarEvents   []RadarEvent      `json:"radar_events,omitempty"`
	SLSDetections []SLSDetection    `json:"sls_detections,omitempty"`
	Interactions  []UserInteraction `json:"interactions,omitempty"`
}

// Location represents geographic and descriptive location data
type Location struct {
	Latitude    float64 `json:"latitude" db:"latitude"`
	Longitude   float64 `json:"longitude" db:"longitude"`
	Address     string  `json:"address" db:"address"`
	Description string  `json:"description" db:"description"`
	Venue       string  `json:"venue" db:"venue"`
}

// Environmental represents environmental conditions during investigation
type Environmental struct {
	Temperature float64 `json:"temperature" db:"temperature"`
	Humidity    float64 `json:"humidity" db:"humidity"`
	Pressure    float64 `json:"pressure" db:"pressure"`
	EMFLevel    float64 `json:"emf_level" db:"emf_level"`
	LightLevel  float64 `json:"light_level" db:"light_level"`
	NoiseLevel  float64 `json:"noise_level" db:"noise_level"`
}

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusPaused   SessionStatus = "paused"
	SessionStatusComplete SessionStatus = "complete"
	SessionStatusArchived SessionStatus = "archived"
)

// EVPRecording represents an Electronic Voice Phenomenon recording
type EVPRecording struct {
	ID             string     `json:"id" db:"id"`
	SessionID      string     `json:"session_id" db:"session_id"`
	FilePath       string     `json:"file_path" db:"file_path"`
	Duration       float64    `json:"duration" db:"duration"`
	Timestamp      time.Time  `json:"timestamp" db:"timestamp"`
	WaveformData   []float64  `json:"waveform_data" db:"waveform_data"`
	ProcessedPath  string     `json:"processed_path,omitempty" db:"processed_path"`
	Annotations    []string   `json:"annotations" db:"annotations"`
	Quality        EVPQuality `json:"quality" db:"quality"`
	DetectionLevel float64    `json:"detection_level" db:"detection_level"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// EVPQuality represents the quality rating of an EVP recording
type EVPQuality string

const (
	EVPQualityExcellent EVPQuality = "excellent"
	EVPQualityGood      EVPQuality = "good"
	EVPQualityFair      EVPQuality = "fair"
	EVPQualityPoor      EVPQuality = "poor"
)

// VOXEvent represents a Voice Synthesis (VOX) communication event
type VOXEvent struct {
	ID              string    `json:"id" db:"id"`
	SessionID       string    `json:"session_id" db:"session_id"`
	Timestamp       time.Time `json:"timestamp" db:"timestamp"`
	GeneratedText   string    `json:"generated_text" db:"generated_text"`
	PhoneticBank    string    `json:"phonetic_bank" db:"phonetic_bank"`
	FrequencyData   []float64 `json:"frequency_data" db:"frequency_data"`
	TriggerStrength float64   `json:"trigger_strength" db:"trigger_strength"`
	LanguagePack    string    `json:"language_pack" db:"language_pack"`
	ModulationType  string    `json:"modulation_type" db:"modulation_type"`
	UserResponse    string    `json:"user_response,omitempty" db:"user_response"`
	ResponseDelay   float64   `json:"response_delay,omitempty" db:"response_delay"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// RadarEvent represents a radar detection event
type RadarEvent struct {
	ID            string        `json:"id" db:"id"`
	SessionID     string        `json:"session_id" db:"session_id"`
	Timestamp     time.Time     `json:"timestamp" db:"timestamp"`
	Position      Coordinates   `json:"position" db:"position"`
	Strength      float64       `json:"strength" db:"strength"`
	SourceType    SourceType    `json:"source_type" db:"source_type"`
	EMFReading    float64       `json:"emf_reading" db:"emf_reading"`
	AudioAnomaly  float64       `json:"audio_anomaly" db:"audio_anomaly"`
	Duration      float64       `json:"duration" db:"duration"`
	MovementTrail []Coordinates `json:"movement_trail,omitempty" db:"movement_trail"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
}

// SLSDetection represents Structured Light Sensor detection data
type SLSDetection struct {
	ID             string           `json:"id" db:"id"`
	SessionID      string           `json:"session_id" db:"session_id"`
	Timestamp      time.Time        `json:"timestamp" db:"timestamp"`
	SkeletalPoints []SkeletalPoint  `json:"skeletal_points" db:"skeletal_points"`
	Confidence     float64          `json:"confidence" db:"confidence"`
	BoundingBox    BoundingBox      `json:"bounding_box" db:"bounding_box"`
	VideoFrame     string           `json:"video_frame,omitempty" db:"video_frame"`
	FilterApplied  []string         `json:"filter_applied" db:"filter_applied"`
	Duration       float64          `json:"duration" db:"duration"`
	Movement       MovementAnalysis `json:"movement" db:"movement"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
}

// UserInteraction represents user interactions during investigation
type UserInteraction struct {
	ID               string            `json:"id" db:"id"`
	SessionID        string            `json:"session_id" db:"session_id"`
	Timestamp        time.Time         `json:"timestamp" db:"timestamp"`
	Type             InteractionType   `json:"type" db:"type"`
	Content          string            `json:"content" db:"content"`
	AudioPath        string            `json:"audio_path,omitempty" db:"audio_path"`
	Response         string            `json:"response,omitempty" db:"response"`
	ResponseTime     float64           `json:"response_time,omitempty" db:"response_time"`
	RandomizerResult *RandomizerResult `json:"randomizer_result,omitempty" db:"randomizer_result"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
}

// Supporting types
type Coordinates struct {
	X float64 `json:"x" db:"x"`
	Y float64 `json:"y" db:"y"`
	Z float64 `json:"z,omitempty" db:"z"`
}

type SourceType string

const (
	SourceTypeEMF   SourceType = "emf"
	SourceTypeAudio SourceType = "audio"
	SourceTypeBoth  SourceType = "both"
	SourceTypeOther SourceType = "other"
)

type SkeletalPoint struct {
	Joint      string      `json:"joint" db:"joint"`
	Position   Coordinates `json:"position" db:"position"`
	Confidence float64     `json:"confidence" db:"confidence"`
}

type BoundingBox struct {
	TopLeft     Coordinates `json:"top_left" db:"top_left"`
	BottomRight Coordinates `json:"bottom_right" db:"bottom_right"`
	Width       float64     `json:"width" db:"width"`
	Height      float64     `json:"height" db:"height"`
}

type MovementAnalysis struct {
	Speed     float64 `json:"speed" db:"speed"`
	Direction float64 `json:"direction" db:"direction"`
	Pattern   string  `json:"pattern" db:"pattern"`
}

type InteractionType string

const (
	InteractionTypeVoice      InteractionType = "voice"
	InteractionTypeText       InteractionType = "text"
	InteractionTypeRandomizer InteractionType = "randomizer"
	InteractionTypeNote       InteractionType = "note"
)

type RandomizerResult struct {
	Type   string      `json:"type" db:"type"`
	Result interface{} `json:"result" db:"result"`
	Range  string      `json:"range" db:"range"`
}
