# Test Infrastructure & Tooling

## Description
Set up comprehensive testing infrastructure, CI/CD pipelines, and automated quality gates for the OtherSide testing framework.

## Tasks

### 1. Testing Dependencies & Tools
- [ ] **Update go.mod with testing dependencies**
   ```go
   // Add to go.mod
   require (
       github.com/stretchr/testify v1.9.0
       github.com/golang/mock v1.6.0
       github.com/DATA-DOG/go-sqlmock v1.5.0
       github.com/gorilla/mux v1.8.1
   )
   ```

- [ ] **Install and configure mock generation tools**
   ```bash
   go install github.com/golang/mock/mockgen@latest
   go install github.com/golang/mock/mockgen@latest
   ```

### 2. Test Utilities & Helpers
- [ ] **Create test helpers package** (`testutils/helpers.go`)
  ```go
  package testutils
  
  func SetupTestDB(t *testing.T) *sql.DB
  func CleanupTestDB(db *sql.DB)
  func GenerateTestSession() *domain.Session
  func GenerateTestAudioData() []float64
  func AssertJSONEqual(t *testing.T, expected, actual interface{})
  func SetupMockServices() (*MockSessionService, *MockAudioProcessor, ...)
  ```

- [ ] **Create test data fixtures** (`testutils/fixtures/`)
  - `test_sessions.json` - Sample session data
  - `test_audio.json` - Audio processing test data
  - `test_events.json` - EVP/VOX/Radar/SLS event samples
  - `test_config.json` - Configuration test scenarios

- [ ] **Create mock generation scripts**
  ```bash
  #!/bin/bash
  # generate_mocks.sh
  mockgen -source=internal/domain/repository.go -destination=testutils/mocks/repository_mock.go
  mockgen -source=pkg/audio/processor.go -destination=testutils/mocks/audio_mock.go
  ```

### 3. Enhanced Makefile Test Commands
- [ ] **Update Makefile test targets**
  ```makefile
  # Enhanced test commands
  test-unit:
  	@echo "Running unit tests..."
  	go test -v -race -coverprofile=coverage.out ./internal/... ./pkg/...
  
  test-integration:
  	@echo "Running integration tests..."
  	go test -v -tags=integration ./test/integration/...
  
  test-all: test-unit test-integration
  
  test-coverage: test-unit
  	@echo "Generating coverage report..."
  	go tool cover -html=coverage.out -o coverage.html
  	go tool cover -func=coverage.out | grep "total:"
  
  benchmark:
  	@echo "Running benchmarks..."
  	go test -bench=. -benchmem -run=^$$ ./pkg/audio/...
  
  test-performance:
  	@echo "Running performance tests..."
  	go test -v -tags=performance ./test/performance/...
  ```

### 4. Continuous Integration Setup
- [ ] **GitHub Actions workflow** (`.github/workflows/test.yml`)
  ```yaml
  name: Test Suite
  on: [push, pull_request]
  
  jobs:
    test:
      runs-on: ubuntu-latest
      strategy:
        matrix:
          go-version: [1.22, 1.23]
      
      steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      
      - name: Install dependencies
        run: |
          go mod download
          go install github.com/golang/mock/mockgen@latest
      
      - name: Generate mocks
        run: ./scripts/generate_mocks.sh
      
      - name: Run tests
        run: make test-all
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
  ```

### 5. Code Quality Gates
- [ ] **Test coverage thresholds**
  - Minimum 80% coverage to pass
  - Target 85% coverage
  - Critical paths 95% coverage

- [ ] **Performance benchmarks**
  - Audio processing <100ms for 1s audio
  - Database queries <50ms average
  - API responses <200ms for 95% of requests

- [ ] **Code quality checks**
  ```bash
  # Add to Makefile
  lint:
  	@echo "Running linters..."
  	golangci-lint run
  
  security:
  	@echo "Running security scan..."
  	gosec ./...
  
  vuln-check:
  	@echo "Checking for vulnerabilities..."
  	govulncheck ./...
  ```

### 6. Test Documentation
- [ ] **Create testing guide** (`docs/testing.md`)
  - How to run tests
  - Test organization
  - Mock usage guidelines
  - Coverage requirements
  - Performance benchmarking

- [ ] **Add test examples** (`docs/test-examples.md`)
  - Domain model testing patterns
  - Service layer testing with mocks
  - HTTP handler testing
  - Repository testing with test databases

### 7. Mock Infrastructure
- [ ] **Repository mocks** (`testutils/mocks/repository_mock.go`)
  - MockSessionRepository
  - MockEVPRepository
  - MockVOXRepository
  - MockRadarRepository
  - MockSLSRepository
  - MockInteractionRepository

- [ ] **Service mocks** (`testutils/mocks/service_mock.go`)
  - MockAudioProcessor
  - MockVOXGenerator
  - MockFileManager

## Test Environment Setup
- [ ] **Docker test environment** (`docker-compose.test.yml`)
  ```yaml
  version: '3.8'
  services:
    test-db:
      image: sqlite:latest
      volumes:
        - ./test/data:/data
      environment:
        - SQLITE_DB=/data/test.db
  ```

- [ ] **Test database initialization**
  - Automated schema setup
  - Test data seeding
  - Cleanup between test runs

## Acceptance Criteria
- [ ] All testing dependencies installed and configured
- [ ] Mock generation working for all interfaces
- [ ] CI/CD pipeline passing with coverage reports
- [ ] Performance benchmarks automated
- [ ] Code quality gates enforcing standards
- [ ] Documentation comprehensive and up-to-date
- [ ] Local test environment easily reproducible

## Labels
`testing`, `infrastructure`, `ci-cd`, `tooling`
`enhancement` `testing` `infrastructure` `ci-cd` `automation` `tooling`

## Priority
High - Foundation for all test implementation
## Estimated Effort
2-3 days
## Dependencies
- GitHub repository access for workflow setup
- Team approval for CI/CD configuration