# Phase 3: Service Layer Tests

## Description
Implement comprehensive unit tests for business logic in service layer. This phase validates paranormal investigation workflows, session management, and data orchestration.

## Tasks

### 1. Session Service Tests (`internal/service/session_test.go`)
- [ ] **Session Creation Tests**
  - Valid session creation with complete data
  - Required field validation (title, location)
  - Environmental data validation
  - ID generation uniqueness
  - Timestamp accuracy (start_time, created_at, updated_at)
  - Default status assignment (active)

- [ ] **EVP Processing Tests**
  - Valid EVP recording workflow
  - Session existence and active status validation
  - Audio processor integration with mocks
  - Quality determination logic testing
  - Repository save operation validation
  - Error handling for invalid session states

- [ ] **VOX Communication Tests**
  - Trigger data validation and preparation
  - VOX generator integration with mocks
  - Threshold-based generation logic
  - Language pack and phonetic bank selection
  - Repository save operation validation
  - Nil result handling for below-threshold triggers

- [ ] **Radar Event Processing Tests**
  - Radar event data validation
  - Validation logic for minimizing false positives
  - Source type determination (emf, audio, both, other)
  - Movement trail validation
  - Repository operations with mocked dependencies

- [ ] **SLS Detection Tests**
  - SLS detection data validation
  - False-positive reduction filters
  - Movement pattern analysis
  - Skeletal point validation
  - Bounding box and confidence validation

- [ ] **User Interaction Tests**
  - Interaction type validation
  - Response time handling
  - Audio path validation
  - Randomizer result processing
  - Repository save operations

- [ ] **Session Summary Tests**
  - Complete data aggregation from all repositories
  - Statistics calculation accuracy
  - EVP quality distribution
  - Average anomaly strength computation
  - Error handling for missing session

### 2. Export Service Tests (`internal/service/export_test.go`)
- [ ] **Data Export Tests**
  - JSON export format validation
  - CSV export generation
  - ZIP file creation with multiple formats
  - Session filtering and selection
  - File naming conventions
  - Export metadata generation

- [ ] **Export List Tests**
  - Export file listing functionality
  - Pagination handling
  - File metadata accuracy
  - Sorting and filtering

- [ ] **File Download Tests**
  - File existence validation
  - Path security checks
  - Content type headers
  - Error handling for missing files

### 3. Session State Service Tests (`internal/service/session_state_test.go`)
- [ ] **State Transition Tests**
  - Valid transition validation
  - Invalid transition rejection
  - Concurrent state change handling
  - State persistence validation
  - Transition audit logging

## Mock Requirements
```go
// Mock interfaces needed
type MockSessionRepository struct {
    mock.Mock
}

type MockEVPRepository struct {
    mock.Mock
}

type MockVOXRepository struct {
    mock.Mock
}

type MockAudioProcessor struct {
    mock.Mock
}

type MockVOXGenerator struct {
    mock.Mock
}
```

## Test Data Fixtures
- Complete session objects
- Audio processing results
- Radar event data samples
- SLS detection samples
- User interaction examples
- Export data sets

## Acceptance Criteria
- [ ] All service methods have 95%+ line coverage
- [ ] Business logic workflows fully tested
- [ ] Error scenarios covered with proper validation
- [ ] Mock dependencies properly verified
- [ ] Complex calculations validated
- [ ] Concurrent access tested where applicable

## Technical Implementation
```go
func TestSessionService_CreateSession_ValidInput_Success(t *testing.T) {
    // Arrange
    mockRepo := &MockSessionRepository{}
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Session")).
        Return(nil).
        Once()
    
    service := NewSessionService(mockRepo, nil, nil, nil, nil, nil, nil, nil, nil)
    
    req := CreateSessionRequest{
        Title: "Test Investigation",
        Location: domain.Location{
            Latitude:  37.7749,
            Longitude: -122.4194,
            Address:   "Test Location",
        },
        Environmental: domain.Environmental{
            Temperature: 20.5,
            Humidity:    45.0,
            Pressure:    1013.25,
        },
    }
    
    // Act
    session, err := service.CreateSession(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, session.ID)
    assert.Equal(t, req.Title, session.Title)
    assert.Equal(t, domain.SessionStatusActive, session.Status)
    assert.WithinDuration(t, time.Now(), session.CreatedAt, time.Second)
    assert.WithinDuration(t, time.Now(), session.StartTime, time.Second)
    
    mockRepo.AssertExpectations(t)
}
```

## Integration Points
- Repository layer mock validation
- Audio processing mock verification
- File system mock integration
- Context cancellation testing
- Error propagation validation

## Labels
`testing`, `service-layer`, `business-logic`, `phase-3`
`enhancement` `testing` `service` `business-logic` `workflows` `phase-3`

## Priority
High - Core business logic validation
## Estimated Effort
4-5 days
## Dependencies
- Phase 1: Domain & Configuration Tests completed
- Phase 2: Audio Processing Tests completed
- Mock infrastructure ready