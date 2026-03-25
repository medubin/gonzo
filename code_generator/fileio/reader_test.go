package fileio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/medubin/gonzo/code_generator/fileio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFile(t *testing.T) {
	t.Run("successful file reading", func(t *testing.T) {
		// Create a temporary file with test content
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := "Hello, World!\nThis is a test file.\n"
		
		err := os.WriteFile(testFile, []byte(testContent), 0644)
		require.NoError(t, err)
		
		// Test ParseFile
		result, err := fileio.ParseFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, []byte(testContent), result)
		assert.Equal(t, testContent, string(result))
	})

	t.Run("empty file reading", func(t *testing.T) {
		// Create a temporary empty file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "empty.txt")
		
		err := os.WriteFile(testFile, []byte{}, 0644)
		require.NoError(t, err)
		
		// Test ParseFile
		result, err := fileio.ParseFile(testFile)
		require.NoError(t, err)
		assert.Empty(t, result)
		assert.Equal(t, []byte{}, result)
	})

	t.Run("binary file reading", func(t *testing.T) {
		// Create a temporary file with binary content
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "binary.dat")
		binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
		
		err := os.WriteFile(testFile, binaryContent, 0644)
		require.NoError(t, err)
		
		// Test ParseFile
		result, err := fileio.ParseFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, binaryContent, result)
	})

	t.Run("large file reading", func(t *testing.T) {
		// Create a temporary file with larger content
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "large.txt")
		
		// Create content larger than typical buffer sizes
		largeContent := make([]byte, 10*1024) // 10KB
		for i := range largeContent {
			largeContent[i] = byte('A' + (i % 26))
		}
		
		err := os.WriteFile(testFile, largeContent, 0644)
		require.NoError(t, err)
		
		// Test ParseFile
		result, err := fileio.ParseFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, largeContent, result)
		assert.Len(t, result, 10*1024)
	})

	t.Run("non-existent file error", func(t *testing.T) {
		nonExistentFile := "/path/that/does/not/exist/file.txt"
		
		result, err := fileio.ParseFile(nonExistentFile)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("permission denied error", func(t *testing.T) {
		// Skip on Windows as permission handling is different
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping permission test on Windows")
		}
		
		// Create a temporary file and remove read permissions
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "noaccess.txt")
		
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		require.NoError(t, err)
		
		// Remove read permissions
		err = os.Chmod(testFile, 0000)
		require.NoError(t, err)
		
		// Restore permissions after test for cleanup
		t.Cleanup(func() {
			os.Chmod(testFile, 0644)
		})
		
		// Test ParseFile
		result, err := fileio.ParseFile(testFile)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("API file parsing", func(t *testing.T) {
		// Test with actual API file format (matches usage in codebase)
		tempDir := t.TempDir()
		apiFile := filepath.Join(tempDir, "test.api")
		apiContent := `// Test API definition
enum UserRole string {
    ADMIN = "admin"
    USER = "user"
}

type User struct {
    id required UserID
    username required string
    role UserRole
}`
		
		err := os.WriteFile(apiFile, []byte(apiContent), 0644)
		require.NoError(t, err)
		
		// Test ParseFile
		result, err := fileio.ParseFile(apiFile)
		require.NoError(t, err)
		
		// Use snapshot test to verify the parsed content structure
		snaps.MatchSnapshot(t, string(result))
	})
}