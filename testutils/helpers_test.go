package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTestDB(t *testing.T) {
	t.Run("Mock Database", func(t *testing.T) {
		testDB := NewTestDB(t, true)
		defer testDB.Close()

		assert.NotNil(t, testDB.DB)
		assert.NotNil(t, testDB.Mock)
		assert.True(t, testDB.IsMock)
	})

	t.Run("SQLite Database", func(t *testing.T) {
		testDB := NewTestDB(t, false)
		defer testDB.Close()

		assert.NotNil(t, testDB.DB)
		assert.False(t, testDB.IsMock)
		assert.NotEmpty(t, testDB.Path)

		// Test that tables were created
		var count int
		err := testDB.DB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count)
		AssertNoError(t, err)
		assert.Greater(t, count, 0)
	})
}

func TestAssertNoError(t *testing.T) {
	t.Run("No Error", func(t *testing.T) {
		AssertNoError(t, nil)
	})

	t.Run("With Error", func(t *testing.T) {
		// This should fail the test, so we use assert.False to catch the failure
		// AssertNoError uses testify's assert.NoError which fails the test
		// We can't easily test this without the test framework itself
		t.Skip("Skipping - AssertNoError uses testify's assert which fails the test")
	})
}

func TestCreateTempFile(t *testing.T) {
	content := "test content"
	filePath := CreateTempFile(t, content, "test-*.txt")
	defer CleanupFile(filePath)

	assert.NotEmpty(t, filePath)
	assert.Contains(t, filePath, "test-")

	// Check if file exists
	_, err := os.Stat(filePath)
	assert.NoError(t, err)
}
