package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HTTPTestHelper provides utilities for HTTP handler testing
type HTTPTestHelper struct {
	T      *testing.T
	Router http.Handler
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(t *testing.T, router http.Handler) *HTTPTestHelper {
	return &HTTPTestHelper{
		T:      t,
		Router: router,
	}
}

// Request performs an HTTP request and returns the response
func (h *HTTPTestHelper) Request(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer

	if body != nil {
		jsonBody, err := json.Marshal(body)
		AssertNoError(h.T, err)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req := httptest.NewRequest(method, path, reqBody)

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default content type for JSON bodies
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	h.Router.ServeHTTP(rr, req)

	return rr
}

// AssertStatus asserts the HTTP response status code
func (h *HTTPTestHelper) AssertStatus(rr *httptest.ResponseRecorder, expected int) {
	assert.Equal(h.T, expected, rr.Code, "Unexpected status code")
}

// AssertJSON asserts the response body contains valid JSON
func (h *HTTPTestHelper) AssertJSON(rr *httptest.ResponseRecorder) {
	var response interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	AssertNoError(h.T, err, "Response should be valid JSON")
}

// AssertJSONResponse asserts the response body matches expected JSON
func (h *HTTPTestHelper) AssertJSONResponse(rr *httptest.ResponseRecorder, expected interface{}) {
	h.AssertJSON(rr)

	var actual interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &actual)
	AssertNoError(h.T, err)

	assert.Equal(h.T, expected, actual, "JSON response mismatch")
}

// AssertHeader asserts a response header value
func (h *HTTPTestHelper) AssertHeader(rr *httptest.ResponseRecorder, key, expected string) {
	actual := rr.Header().Get(key)
	assert.Equal(h.T, expected, actual, "Header %s mismatch", key)
}

// AssertContentType asserts the content type header
func (h *HTTPTestHelper) AssertContentType(rr *httptest.ResponseRecorder, expected string) {
	h.AssertHeader(rr, "Content-Type", expected)
}

// GetJSONBody unmarshals JSON response body into target
func (h *HTTPTestHelper) GetJSONBody(rr *httptest.ResponseRecorder, target interface{}) {
	err := json.Unmarshal(rr.Body.Bytes(), target)
	AssertNoError(h.T, err, "Failed to unmarshal JSON response")
}

// TableTest represents a table-driven test case
type TableTest struct {
	Name     string
	Input    interface{}
	Expected interface{}
	Error    bool
}

// RunTableTests executes table-driven tests
func RunTableTests(t *testing.T, tests []TableTest, testFunc func(t *testing.T, tt TableTest)) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testFunc(t, tt)
		})
	}
}
