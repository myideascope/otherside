package service

import (
	"context"
	"time"

	"github.com/myideascope/otherside/internal/domain"
	"github.com/myideascope/otherside/pkg/audio"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository mocks SessionRepository interface
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Session, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func (m *MockSessionRepository) GetByStatus(ctx context.Context, status domain.SessionStatus) ([]*domain.Session, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByDateRange(ctx context.Context, start, end time.Time) ([]*domain.Session, error) {
	args := m.Called(ctx, start, end)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

// MockEVPRepository mocks EVPRepository interface
type MockEVPRepository struct {
	mock.Mock
}

func (m *MockEVPRepository) Create(ctx context.Context, evp *domain.EVPRecording) error {
	args := m.Called(ctx, evp)
	return args.Error(0)
}

func (m *MockEVPRepository) GetByID(ctx context.Context, id string) (*domain.EVPRecording, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.EVPRecording), args.Error(1)
}

func (m *MockEVPRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.EVPRecording, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.EVPRecording), args.Error(1)
}

func (m *MockEVPRepository) Update(ctx context.Context, evp *domain.EVPRecording) error {
	args := m.Called(ctx, evp)
	return args.Error(0)
}

func (m *MockEVPRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEVPRepository) GetByQuality(ctx context.Context, quality domain.EVPQuality) ([]*domain.EVPRecording, error) {
	args := m.Called(ctx, quality)
	return args.Get(0).([]*domain.EVPRecording), args.Error(1)
}

func (m *MockEVPRepository) GetByDetectionLevel(ctx context.Context, minLevel float64) ([]*domain.EVPRecording, error) {
	args := m.Called(ctx, minLevel)
	return args.Get(0).([]*domain.EVPRecording), args.Error(1)
}

// MockVOXRepository mocks VOXRepository interface
type MockVOXRepository struct {
	mock.Mock
}

func (m *MockVOXRepository) Create(ctx context.Context, vox *domain.VOXEvent) error {
	args := m.Called(ctx, vox)
	return args.Error(0)
}

func (m *MockVOXRepository) GetByID(ctx context.Context, id string) (*domain.VOXEvent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.VOXEvent), args.Error(1)
}

func (m *MockVOXRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.VOXEvent, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.VOXEvent), args.Error(1)
}

func (m *MockVOXRepository) Update(ctx context.Context, vox *domain.VOXEvent) error {
	args := m.Called(ctx, vox)
	return args.Error(0)
}

func (m *MockVOXRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVOXRepository) GetByLanguagePack(ctx context.Context, languagePack string) ([]*domain.VOXEvent, error) {
	args := m.Called(ctx, languagePack)
	return args.Get(0).([]*domain.VOXEvent), args.Error(1)
}

func (m *MockVOXRepository) GetByTriggerStrength(ctx context.Context, minStrength float64) ([]*domain.VOXEvent, error) {
	args := m.Called(ctx, minStrength)
	return args.Get(0).([]*domain.VOXEvent), args.Error(1)
}

// MockRadarRepository mocks RadarRepository interface
type MockRadarRepository struct {
	mock.Mock
}

func (m *MockRadarRepository) Create(ctx context.Context, radar *domain.RadarEvent) error {
	args := m.Called(ctx, radar)
	return args.Error(0)
}

func (m *MockRadarRepository) GetByID(ctx context.Context, id string) (*domain.RadarEvent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.RadarEvent), args.Error(1)
}

func (m *MockRadarRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.RadarEvent, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.RadarEvent), args.Error(1)
}

func (m *MockRadarRepository) Update(ctx context.Context, radar *domain.RadarEvent) error {
	args := m.Called(ctx, radar)
	return args.Error(0)
}

func (m *MockRadarRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRadarRepository) GetBySourceType(ctx context.Context, sourceType domain.SourceType) ([]*domain.RadarEvent, error) {
	args := m.Called(ctx, sourceType)
	return args.Get(0).([]*domain.RadarEvent), args.Error(1)
}

func (m *MockRadarRepository) GetByStrengthRange(ctx context.Context, minStrength, maxStrength float64) ([]*domain.RadarEvent, error) {
	args := m.Called(ctx, minStrength, maxStrength)
	return args.Get(0).([]*domain.RadarEvent), args.Error(1)
}

// MockSLSRepository mocks SLSRepository interface
type MockSLSRepository struct {
	mock.Mock
}

func (m *MockSLSRepository) Create(ctx context.Context, sls *domain.SLSDetection) error {
	args := m.Called(ctx, sls)
	return args.Error(0)
}

func (m *MockSLSRepository) GetByID(ctx context.Context, id string) (*domain.SLSDetection, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.SLSDetection), args.Error(1)
}

func (m *MockSLSRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.SLSDetection, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.SLSDetection), args.Error(1)
}

func (m *MockSLSRepository) Update(ctx context.Context, sls *domain.SLSDetection) error {
	args := m.Called(ctx, sls)
	return args.Error(0)
}

func (m *MockSLSRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSLSRepository) GetByConfidenceRange(ctx context.Context, minConfidence float64) ([]*domain.SLSDetection, error) {
	args := m.Called(ctx, minConfidence)
	return args.Get(0).([]*domain.SLSDetection), args.Error(1)
}

func (m *MockSLSRepository) GetByDuration(ctx context.Context, minDuration float64) ([]*domain.SLSDetection, error) {
	args := m.Called(ctx, minDuration)
	return args.Get(0).([]*domain.SLSDetection), args.Error(1)
}

// MockInteractionRepository mocks InteractionRepository interface
type MockInteractionRepository struct {
	mock.Mock
}

func (m *MockInteractionRepository) Create(ctx context.Context, interaction *domain.UserInteraction) error {
	args := m.Called(ctx, interaction)
	return args.Error(0)
}

func (m *MockInteractionRepository) GetByID(ctx context.Context, id string) (*domain.UserInteraction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.UserInteraction), args.Error(1)
}

func (m *MockInteractionRepository) GetBySessionID(ctx context.Context, sessionID string) ([]*domain.UserInteraction, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]*domain.UserInteraction), args.Error(1)
}

func (m *MockInteractionRepository) Update(ctx context.Context, interaction *domain.UserInteraction) error {
	args := m.Called(ctx, interaction)
	return args.Error(0)
}

func (m *MockInteractionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInteractionRepository) GetByType(ctx context.Context, interactionType domain.InteractionType) ([]*domain.UserInteraction, error) {
	args := m.Called(ctx, interactionType)
	return args.Get(0).([]*domain.UserInteraction), args.Error(1)
}

// MockFileRepository mocks FileRepository interface
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) SaveFile(ctx context.Context, path string, data []byte) error {
	args := m.Called(ctx, path, data)
	return args.Error(0)
}

func (m *MockFileRepository) GetFile(ctx context.Context, path string) ([]byte, error) {
	args := m.Called(ctx, path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFileRepository) DeleteFile(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockFileRepository) FileExists(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileRepository) GetFileSize(ctx context.Context, path string) (int64, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileRepository) ListFiles(ctx context.Context, directory string) ([]string, error) {
	args := m.Called(ctx, directory)
	return args.Get(0).([]string), args.Error(1)
}

// MockAudioProcessor mocks the audio.Processor
type MockAudioProcessor struct {
	mock.Mock
}

func (m *MockAudioProcessor) ProcessAudio(ctx context.Context, audioData []float64) (*audio.ProcessingResult, error) {
	args := m.Called(ctx, audioData)
	return args.Get(0).(*audio.ProcessingResult), args.Error(1)
}

// MockVOXGenerator mocks the audio.VOXGenerator
type MockVOXGenerator struct {
	mock.Mock
}

func (m *MockVOXGenerator) GenerateVOX(ctx context.Context, triggers map[string]float64, config audio.VOXConfig) (*audio.VOXResult, error) {
	args := m.Called(ctx, triggers, config)
	return args.Get(0).(*audio.VOXResult), args.Error(1)
}
