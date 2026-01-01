# Phase 1: Domain & Configuration Tests

## Description
Implement comprehensive unit tests for domain models and configuration management. This phase establishes the foundation by testing core business rules and system configuration.

## Tasks

### 1. Domain Model Tests (`internal/domain/`)
- [ ] **Session Model Tests** (`session_test.go`)
  - Valid session creation with all required fields
  - Session status transitions (active → paused → complete → archived)
  - Location data validation (latitude/longitude bounds, required fields)
  - Environmental data bounds checking
  - JSON serialization/deserialization
  - Database tag validation

- [ ] **EVP Recording Model Tests** (`evp_test.go`)
  - Quality level validation (excellent, good, fair, poor)
  - Detection level range validation (0.0-1.0)
  - Annotation handling (empty, null, multiple)
  - Waveform data validation (non-empty, proper format)
  - Timestamp handling

- [ ] **VOX Event Model Tests** (`vox_test.go`)
  - Trigger strength validation (0.0-1.0)
  - Language pack validation
  - Phonetic bank validation
  - Frequency data validation
  - Response delay calculations

- [ ] **Radar Event Model Tests** (`radar_test.go`)
  - Source type validation (emf, audio, both, other)
  - Position coordinate validation
  - Movement trail data validation
  - EMF reading range validation
  - Audio anomaly bounds checking

- [ ] **SLS Detection Model Tests** (`sls_test.go`)
  - Skeletal points validation (minimum count, position bounds)
  - Confidence score range (0.0-1.0)
  - Bounding box validation
  - Movement analysis calculations
  - Filter application validation

- [ ] **User Interaction Model Tests** (`interaction_test.go`)
  - Interaction type validation (voice, text, randomizer, note)
  - Response time bounds
  - Randomizer result validation
  - Content length validation

### 2. Configuration Tests (`internal/config/`)
- [ ] **Config Loading Tests** (`config_test.go`)
  - Environment variable parsing with valid values
  - Default value fallbacks when env vars missing
  - Type conversion validation (string → int/float)
  - Invalid input handling (malformed numbers, negative values)
  - Required field validation

## Test Data Requirements
- Sample session objects with valid/invalid data
- Environmental data edge cases
- Audio metadata test fixtures
- Invalid configuration scenarios

## Acceptance Criteria
- [ ] All domain models have 90%+ line coverage
- [ ] All validation rules tested with edge cases
- [ ] Configuration loading handles all scenarios
- [ ] Table-driven tests for complex validation
- [ ] Proper error messages for invalid inputs
- [ ] JSON marshaling/unmarshaling tested

## Technical Implementation
```go
// Example table-driven test structure
func TestSessionStatus_Transitions(t *testing.T) {
    tests := []struct {
        name     string
        from     SessionStatus
        to       SessionStatus
        expected bool
    }{
        {"ActiveToPaused", SessionStatusActive, SessionStatusPaused, true},
        {"PausedToActive", SessionStatusPaused, SessionStatusActive, true},
        {"ArchivedToActive", SessionStatusArchived, SessionStatusActive, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isValidTransition(tt.from, tt.to)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Labels
`testing`, `domain`, `configuration`, `phase-1`
`enhancement` `testing` `domain` `validation` `phase-1`

## Priority
High - Foundation for all other testing phases
## Estimated Effort
3-4 days