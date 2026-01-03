# Phase 5: Handler Layer Tests

## Description
Implement comprehensive unit tests for HTTP API endpoints in handler layer. This phase validates request/response handling, status codes, error responses, and API contract compliance.

## Tasks

### 1. Session Handlers Tests (`internal/handler/session_test.go`)
- [ ] **Create Session Endpoint Tests**
  - Valid session creation request
  - Required field validation (title missing)
  - Invalid data validation (invalid coordinates, malformed environmental data)
  - JSON parsing error handling
  - Success response with proper location header
  - Database error propagation

- [ ] **Get Session by ID Tests**
  - Valid session ID retrieval
  - Non-existent session handling (404)
  - Invalid ID format handling (400)
  - Response structure validation
  - Include related data flag handling

- [ ] **List Sessions Tests**
  - Pagination parameters (limit, offset)
  - Status filtering query parameters
  - Date range filtering
  - Sorting parameter validation
  - Empty result handling
  - Invalid query parameter handling

- [ ] **Update Session Tests**
  - Partial update functionality
  - Status transition validation
  - Invalid update rejection
  - Concurrent update handling
  - Optimistic locking (if implemented)

- [ ] **Delete Session Tests**
  - Successful deletion
  - Non-existent session handling
  - Cascade deletion verification
  - Permission checks (if implemented)

### 2. EVP Recording Handlers Tests (`internal/handler/evp_test.go`)
- [ ] **Process EVP Endpoint Tests**
  - Valid audio data upload
  - File size and format validation
  - Session existence validation
  - Processing result response format
  - Async processing status (if implemented)
  - Error handling for corrupted audio

### 3. VOX Communication Handlers Tests (`internal/handler/vox_test.go`)
- [ ] **Generate VOX Endpoint Tests**
  - Valid trigger data processing
  - Environmental data validation
  - VOX result response format
  - Threshold-based no-response handling
  - Language pack validation

### 4. Radar Event Handlers Tests (`internal/handler/radar_test.go`)
- [ ] **Process Radar Event Tests**
  - Valid radar data submission
  - Coordinate validation
  - Source type validation
  - Movement trail data handling
  - False positive reduction validation

### 5. SLS Detection Handlers Tests (`internal/handler/sls_test.go`)
- [ ] **Process SLS Detection Tests**
  - Valid SLS data upload
  - Skeletal points validation
  - Confidence score validation
  - Bounding box validation
  - Movement analysis response

### 6. User Interaction Handlers Tests (`internal/handler/interaction_test.go`)
- [ ] **Record Interaction Tests**
  - Voice interaction recording
  - Text interaction submission
  - Randomizer result handling
  - Audio file upload validation
  - Response correlation

### 7. Export Handlers Tests (`internal/handler/export_test.go`)
- [ ] **Export Sessions Tests**
  - Export format selection (JSON, CSV, ZIP)
  - Session filter application
  - Export job initiation
  - Progress tracking (if async)
  - File generation validation

- [ ] **List Exports Tests**
  - Export listing functionality
  - File metadata accuracy
  - Pagination handling
  - Sorting options

- [ ] **Download Export Tests**
  - File download validation
  - Content type headers
  - File existence verification
  - Download authorization
  - File cleanup after download

## HTTP Testing Setup
```go
func setupTestRouter() *mux.Router {
    router := mux.NewRouter()
    
    // Register handlers
    sessionHandler := NewSessionHandler(mockService, mockConfig)
    router.HandleFunc("/api/v1/sessions", sessionHandler.CreateSession).Methods("POST")
    router.HandleFunc("/api/v1/sessions/{id}", sessionHandler.GetSession).Methods("GET")
    router.HandleFunc("/api/v1/sessions", sessionHandler.ListSessions).Methods("GET")
    router.HandleFunc("/api/v1/sessions/{id}", sessionHandler.UpdateSession).Methods("PUT")
    router.HandleFunc("/api/v1/sessions/{id}", sessionHandler.DeleteSession).Methods("DELETE")
    
    return router
}

func TestSessionHandler_CreateSession_ValidRequest_Success(t *testing.T) {
    // Arrange
    router := setupTestRouter()
    mockService := &MockSessionService{}
    
    reqBody := map[string]interface{}{
        "title": "Test Investigation",
        "location": map[string]interface{}{
            "latitude":  37.7749,
            "longitude": -122.4194,
            "address":   "Test Location",
        },
        "environmental": map[string]interface{}{
            "temperature": 20.5,
            "humidity":    45.0,
            "pressure":    1013.25,
        },
    }
    
    expectedSession := &domain.Session{
        ID:    "generated-id",
        Title: "Test Investigation",
        Status: domain.SessionStatusActive,
    }
    
    mockService.On("CreateSession", mock.Anything, mock.Anything).
        Return(expectedSession, nil).
        Once()
    
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/v1/sessions", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    
    // Act
    router.ServeHTTP(rr, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, rr.Code)
    assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
    
    var response map[string]interface{}
    err := json.Unmarshal(rr.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, expectedSession.ID, response["id"])
    assert.Equal(t, expectedSession.Title, response["title"])
    
    mockService.AssertExpectations(t)
}
```

## Request/Response Validation Tests
- [ ] JSON schema validation
- [ ] Content-Type header validation
- [ ] CORS headers testing
- [ ] Compression support (if implemented)
- [ ] API versioning validation

## Error Handling Tests
- [ ] 400 Bad Request responses
- [ ] 401 Unauthorized responses
- [ ] 403 Forbidden responses
- [ ] 404 Not Found responses
- [ ] 409 Conflict responses
- [ ] 422 Unprocessable Entity responses
- [ ] 500 Internal Server Error responses
- [ ] Error message format consistency
- [ ] Error code standardization

## Security Tests
- [ ] Input validation for injection attacks
- [ ] File upload security validation
- [ ] Rate limiting validation
- [ ] Request size limits
- [ ] Path traversal prevention

## Performance Tests
- [ ] Response time benchmarks
- [ ] Concurrent request handling
- [ ] Memory usage during high load
- [ ] Large payload handling

## Acceptance Criteria
- [ ] All HTTP endpoints have 80%+ line coverage
- [ ] Request/response contracts fully validated
- [ ] Error scenarios comprehensively tested
- [ ] Security measures verified
- [ ] Performance meets SLA requirements
- [ ] API documentation compliance

## Labels
`testing`, `handlers`, `http-api`, `phase-5`
`enhancement` `testing` `http` `api` `handlers` `security` `phase-5`

## Priority
High - API contract validation and security
## Estimated Effort
3-4 days
## Dependencies
- Phase 1-4: All previous phases completed
- Service layer mocks available
- HTTP test infrastructure ready