# Test Infrastructure Setup

## Description
Add testing dependencies and set up the basic test infrastructure for the OtherSide paranormal investigation application.

## Tasks
1. **Add testing dependencies to go.mod**
   - `github.com/stretchr/testify/assert` - Assertion library
   - `github.com/stretchr/testify/mock` - Mocking framework  
   - `github.com/stretchr/testify/suite` - Test suite organization
   - `github.com/golang/mock/gomock` - Alternative mocking option
   - `github.com/DATA-DOG/go-sqlmock` - Database mocking
   - `github.com/gorilla/mux` (if not already present) - HTTP testing

2. **Create test utilities directory structure**
   ```
   testutils/
   ├── mocks/          # Generated mocks
   ├── fixtures/        # Test data
   └── helpers.go      # Common test helpers
   ```

3. **Create test configuration**
   - Test database setup
   - Mock configuration values
   - Test audio samples

4. **Update Makefile test targets**
   - Enhanced test coverage reporting
   - Benchmark support
   - Test cleanup utilities

## Acceptance Criteria
- [ ] All testing dependencies added to go.mod
- [ ] Test directory structure created
- [ ] Basic test helpers implemented
- [ ] Makefile updated with enhanced test commands
- [ ] Test database setup working
- [ ] Mock generation tools configured

## Technical Notes
- Use table-driven tests where appropriate
- Implement proper cleanup with defer statements
- Ensure test isolation between test cases
- Use testify for assertions and mocking

## Labels
`testing`, `infrastructure`, `dependencies`
`enhancement` `testing` `dependencies` `go.mod` `makefile`

## Priority
High - Foundation for all subsequent test implementation