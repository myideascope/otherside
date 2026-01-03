package service

import (
	"context"
	"testing"
	"time"

	"github.com/myideascope/otherside/internal/domain"
	"github.com/myideascope/otherside/pkg/audio"
	"github.com/stretchr/testify/assert"
)

func TestSessionService_determineEVPQuality_ExcellentQuality(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	result := &audio.ProcessingResult{
		AnomalyStrength: 0.9,
		NoiseLevel:      0.05,
	}

	// Act
	quality := service.determineEVPQuality(result)

	// Assert
	assert.Equal(t, domain.EVPQualityExcellent, quality)
}

func TestSessionService_determineEVPQuality_GoodQuality(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	result := &audio.ProcessingResult{
		AnomalyStrength: 0.7,
		NoiseLevel:      0.15,
	}

	// Act
	quality := service.determineEVPQuality(result)

	// Assert
	assert.Equal(t, domain.EVPQualityGood, quality)
}

func TestSessionService_determineEVPQuality_FairQuality(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	result := &audio.ProcessingResult{
		AnomalyStrength: 0.5,
		NoiseLevel:      0.3,
	}

	// Act
	quality := service.determineEVPQuality(result)

	// Assert
	assert.Equal(t, domain.EVPQualityFair, quality)
}

func TestSessionService_determineEVPQuality_PoorQuality(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	result := &audio.ProcessingResult{
		AnomalyStrength: 0.3,
		NoiseLevel:      0.5,
	}

	// Act
	quality := service.determineEVPQuality(result)

	// Assert
	assert.Equal(t, domain.EVPQualityPoor, quality)
}

func TestSessionService_validateRadarEvent_ValidData_ReturnsTrue(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()

	// Act
	valid := service.validateRadarEvent(data)

	// Assert
	assert.True(t, valid)
}

func TestSessionService_validateRadarEvent_InvalidStrength_ReturnsFalse(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.Strength = 0.1 // Below threshold

	// Act
	valid := service.validateRadarEvent(data)

	// Assert
	assert.False(t, valid)
}

func TestSessionService_validateRadarEvent_InvalidPosition_ReturnsFalse(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.Position = domain.Coordinates{X: 0, Y: 0, Z: 0}

	// Act
	valid := service.validateRadarEvent(data)

	// Assert
	assert.False(t, valid)
}

func TestSessionService_validateRadarEvent_InvalidEMFReading_ReturnsFalse(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.EMFReading = 1500 // Above max threshold

	// Act
	valid := service.validateRadarEvent(data)

	// Assert
	assert.False(t, valid)
}

func TestSessionService_determineRadarSourceType_BothHigh_ReturnsBoth(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.EMFReading = 0.8
	data.AudioAnomaly = 0.8

	// Act
	sourceType := service.determineRadarSourceType(data)

	// Assert
	assert.Equal(t, domain.SourceTypeBoth, sourceType)
}

func TestSessionService_determineRadarSourceType_EMFHigh_ReturnsEMF(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.EMFReading = 0.8
	data.AudioAnomaly = 0.2

	// Act
	sourceType := service.determineRadarSourceType(data)

	// Assert
	assert.Equal(t, domain.SourceTypeEMF, sourceType)
}

func TestSessionService_determineRadarSourceType_AudioHigh_ReturnsAudio(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.EMFReading = 0.2
	data.AudioAnomaly = 0.8

	// Act
	sourceType := service.determineRadarSourceType(data)

	// Assert
	assert.Equal(t, domain.SourceTypeAudio, sourceType)
}

func TestSessionService_determineRadarSourceType_BothLow_ReturnsOther(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestRadarEventData()
	data.EMFReading = 0.2
	data.AudioAnomaly = 0.2

	// Act
	sourceType := service.determineRadarSourceType(data)

	// Assert
	assert.Equal(t, domain.SourceTypeOther, sourceType)
}

func TestSessionService_validateSLSDetection_ValidData_ReturnsTrue(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestSLSDetectionData()

	// Act
	valid := service.validateSLSDetection(data)

	// Assert
	assert.True(t, valid)
}

func TestSessionService_validateSLSDetection_LowConfidence_ReturnsFalse(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestSLSDetectionData()
	data.Confidence = 0.3 // Below threshold

	// Act
	valid := service.validateSLSDetection(data)

	// Assert
	assert.False(t, valid)
}

func TestSessionService_validateSLSDetection_InsufficientPoints_ReturnsFalse(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestSLSDetectionData()
	data.SkeletalPoints = []domain.SkeletalPoint{} // Less than 5 points

	// Act
	valid := service.validateSLSDetection(data)

	// Assert
	assert.False(t, valid)
}

func TestSessionService_validateSLSDetection_InvalidBoundingBox_ReturnsFalse(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	data := TestSLSDetectionData()
	data.BoundingBox = domain.BoundingBox{Width: 5, Height: 5} // Too small

	// Act
	valid := service.validateSLSDetection(data)

	// Assert
	assert.False(t, valid)
}

func TestSessionService_analyzeMovementPattern_NoPoints_ReturnsStatic(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	points := []domain.SkeletalPoint{}

	// Act
	movement := service.analyzeMovementPattern(points)

	// Assert
	assert.Equal(t, 0.0, movement.Speed)
	assert.Equal(t, 0.0, movement.Direction)
	assert.Equal(t, "static", movement.Pattern)
}

func TestSessionService_analyzeMovementPattern_SinglePoint_ReturnsStatic(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	points := []domain.SkeletalPoint{
		{Joint: "head", Position: domain.Coordinates{X: 0, Y: 1.8, Z: 0}, Confidence: 0.9},
	}

	// Act
	movement := service.analyzeMovementPattern(points)

	// Assert
	assert.Equal(t, 0.0, movement.Speed)
	assert.Equal(t, 0.0, movement.Direction)
	assert.Equal(t, "static", movement.Pattern)
}

func TestSessionService_analyzeMovementPattern_LinearMovement_ReturnsLinear(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	points := []domain.SkeletalPoint{
		{Joint: "head", Position: domain.Coordinates{X: 0, Y: 1.8, Z: 0}, Confidence: 0.9},
		{Joint: "head", Position: domain.Coordinates{X: 0.1, Y: 1.8, Z: 0}, Confidence: 0.9},
		{Joint: "head", Position: domain.Coordinates{X: 0.2, Y: 1.8, Z: 0}, Confidence: 0.9},
	}

	// Act
	movement := service.analyzeMovementPattern(points)

	// Assert
	assert.Greater(t, movement.Speed, 0.0)
	assert.Greater(t, movement.Direction, 0.0)
	assert.Equal(t, "linear", movement.Pattern)
}

func TestSessionService_calculateSessionStatistics_EmptyData_ReturnsZeros(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	evps := []*domain.EVPRecording{}
	voxEvents := []*domain.VOXEvent{}
	radarEvents := []*domain.RadarEvent{}
	slsDetections := []*domain.SLSDetection{}
	interactions := []*domain.UserInteraction{}

	// Act
	stats := service.calculateSessionStatistics(evps, voxEvents, radarEvents, slsDetections, interactions)

	// Assert
	assert.Equal(t, 0, stats.TotalEVPs)
	assert.Equal(t, 0, stats.TotalVOXEvents)
	assert.Equal(t, 0, stats.TotalRadarEvents)
	assert.Equal(t, 0, stats.TotalSLSDetections)
	assert.Equal(t, 0, stats.TotalInteractions)
	assert.Equal(t, 0, stats.HighQualityEVPs)
	assert.Equal(t, 0, stats.MediumQualityEVPs)
	assert.Equal(t, 0.0, stats.AverageAnomalyStrength)
}

func TestSessionService_calculateSessionStatistics_MixedQualities_ReturnsCorrectCounts(t *testing.T) {
	// Arrange
	service := NewSessionService(
		nil, nil, nil, nil, nil, nil, nil,
	)

	evps := []*domain.EVPRecording{
		{ID: "1", Quality: domain.EVPQualityExcellent, DetectionLevel: 0.9},
		{ID: "2", Quality: domain.EVPQualityGood, DetectionLevel: 0.7},
		{ID: "3", Quality: domain.EVPQualityPoor, DetectionLevel: 0.3},
	}
	voxEvents := []*domain.VOXEvent{{ID: "1"}}
	radarEvents := []*domain.RadarEvent{{ID: "1"}}
	slsDetections := []*domain.SLSDetection{{ID: "1"}}
	interactions := []*domain.UserInteraction{{ID: "1"}}

	// Act
	stats := service.calculateSessionStatistics(evps, voxEvents, radarEvents, slsDetections, interactions)

	// Assert
	assert.Equal(t, 3, stats.TotalEVPs)
	assert.Equal(t, 1, stats.TotalVOXEvents)
	assert.Equal(t, 1, stats.TotalRadarEvents)
	assert.Equal(t, 1, stats.TotalSLSDetections)
	assert.Equal(t, 1, stats.TotalInteractions)
	assert.Equal(t, 1, stats.HighQualityEVPs)
	assert.Equal(t, 1, stats.MediumQualityEVPs)
	assert.Equal(t, 0.6333333333333333, stats.AverageAnomalyStrength) // (0.9 + 0.7 + 0.3) / 3
}
