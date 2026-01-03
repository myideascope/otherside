package service

import (
	"context"
	"testing"

	"github.com/myideascope/otherside/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExportService_ExportSessions_JSONFormat_Success(t *testing.T) {
	// Arrange
	mockSessionRepo := &MockSessionRepository{}
	mockEVPRepo := &MockEVPRepository{}
	mockVOXRepo := &MockVOXRepository{}
	mockRadarRepo := &MockRadarRepository{}
	mockSLSRepo := &MockSLSRepository{}
	mockInteractionRepo := &MockInteractionRepository{}
	mockFileRepo := &MockFileRepository{}

	testSession := TestSession()
	evps := []*domain.EVPRecording{TestEVPRecording()}

	mockSessionRepo.On("GetByID", mock.Anything, "test-session-123").
		Return(testSession, nil).
		Once()

	mockEVPRepo.On("GetBySessionID", mock.Anything, "test-session-123").
		Return(evps, nil).
		Once()

	mockFileRepo.On("SaveFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
		Return(nil).
		Once()

	service := NewExportService(
		mockSessionRepo, mockEVPRepo, mockVOXRepo, mockRadarRepo,
		mockSLSRepo, mockInteractionRepo, mockFileRepo,
	)

	req := TestExportRequest()
	req.Format = ExportFormatJSON
	req.SessionIDs = []string{"test-session-123"}

	// Act
	result, err := service.ExportSessions(context.Background(), req)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Filename)
	assert.Greater(t, result.Size, int64(0))
	assert.Equal(t, "json", result.Format)
	assert.Equal(t, 1, result.SessionCount)
	assert.Equal(t, 1, result.ItemCount) // 1 EVP
	assert.NotNil(t, result.GeneratedAt)
	assert.NotNil(t, result.FilePath)
	assert.Contains(t, result.Filename, ".json")

	mockSessionRepo.AssertExpectations(t)
	mockEVPRepo.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}

func TestExportService_ExportSessions_EmptySessionIDs_ReturnsError(t *testing.T) {
	// Arrange
	mockSessionRepo := &MockSessionRepository{}
	mockEVPRepo := &MockEVPRepository{}
	mockVOXRepo := &MockVOXRepository{}
	mockRadarRepo := &MockRadarRepository{}
	mockSLSRepo := &MockSLSRepository{}
	mockInteractionRepo := &MockInteractionRepository{}
	mockFileRepo := &MockFileRepository{}

	service := NewExportService(
		mockSessionRepo, mockEVPRepo, mockVOXRepo, mockRadarRepo,
		mockSLSRepo, mockInteractionRepo, mockFileRepo,
	)

	req := TestExportRequest()
	req.SessionIDs = []string{} // Empty

	// Act
	result, err := service.ExportSessions(context.Background(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no sessions specified for export")
}

func TestExportService_ExportSessions_UnsupportedFormat_ReturnsError(t *testing.T) {
	// Arrange
	mockSessionRepo := &MockSessionRepository{}
	mockEVPRepo := &MockEVPRepository{}
	mockVOXRepo := &MockVOXRepository{}
	mockRadarRepo := &MockRadarRepository{}
	mockSLSRepo := &MockSLSRepository{}
	mockInteractionRepo := &MockInteractionRepository{}
	mockFileRepo := &MockFileRepository{}

	service := NewExportService(
		mockSessionRepo, mockEVPRepo, mockVOXRepo, mockRadarRepo,
		mockSLSRepo, mockInteractionRepo, mockFileRepo,
	)

	req := TestExportRequest()
	req.Format = ExportFormat("unsupported")
	req.SessionIDs = []string{"test-session-123"}

	// Act
	result, err := service.ExportSessions(context.Background(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestExportService_ExportSessions_SessionNotFound_ReturnsError(t *testing.T) {
	// Arrange
	mockSessionRepo := &MockSessionRepository{}
	mockEVPRepo := &MockEVPRepository{}
	mockVOXRepo := &MockVOXRepository{}
	mockRadarRepo := &MockRadarRepository{}
	mockSLSRepo := &MockSLSRepository{}
	mockInteractionRepo := &MockInteractionRepository{}
	mockFileRepo := &MockFileRepository{}

	expectedError := assert.AnError
	mockSessionRepo.On("GetByID", mock.Anything, "nonexistent-session").
		Return(nil, expectedError).
		Once()

	service := NewExportService(
		mockSessionRepo, mockEVPRepo, mockVOXRepo, mockRadarRepo,
		mockSLSRepo, mockInteractionRepo, mockFileRepo,
	)

	req := TestExportRequest()
	req.SessionIDs = []string{"nonexistent-session"}

	// Act
	result, err := service.ExportSessions(context.Background(), req)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to collect data for session nonexistent-session")

	mockSessionRepo.AssertExpectations(t)
}
