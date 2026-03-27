package fileio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/medubin/gonzo/code_generator/fileio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteToFile(t *testing.T) {
	t.Run("successful file writing", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "test.go"
		content := "package main\n\nfunc main() {\n    println(\"Hello, World!\")\n}"
		
		err := fileio.WriteToFile(tempDir, fileName, content, false)
		require.NoError(t, err)
		
		// Verify file was created with correct content
		filePath := filepath.Join(tempDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(result))
		
		// Verify file permissions
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
	})

	t.Run("creates single directory if it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()
		subDir := filepath.Join(tempDir, "subdir")
		fileName := "test.txt"
		content := "test content"
		
		// Directory doesn't exist initially
		_, err := os.Stat(subDir)
		assert.True(t, os.IsNotExist(err))
		
		err = fileio.WriteToFile(subDir, fileName, content, false)
		require.NoError(t, err)
		
		// Verify directory was created
		info, err := os.Stat(subDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		
		// Verify file was written
		filePath := filepath.Join(subDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(result))
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedDir := filepath.Join(tempDir, "subdir", "nested")
		fileName := "test.txt"
		content := "test content"

		// Nested directory doesn't exist initially
		_, err := os.Stat(nestedDir)
		assert.True(t, os.IsNotExist(err))

		// WriteToFile should create all intermediate directories
		err = fileio.WriteToFile(nestedDir, fileName, content, false)
		require.NoError(t, err)

		// Verify the file was written
		result, err := os.ReadFile(filepath.Join(nestedDir, fileName))
		require.NoError(t, err)
		assert.Equal(t, content, string(result))
	})

	t.Run("safe mode - doesn't overwrite existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "existing.go"
		originalContent := "// Original content"
		newContent := "// New content"
		
		// Create initial file
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte(originalContent), 0644)
		require.NoError(t, err)
		
		// Try to write with safe mode - should not overwrite
		err = fileio.WriteToFile(tempDir, fileName, newContent, true)
		require.NoError(t, err)
		
		// Verify original content is preserved
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, originalContent, string(result))
		assert.NotEqual(t, newContent, string(result))
	})

	t.Run("safe mode - writes to non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "new.go"
		content := "// New file content"
		
		err := fileio.WriteToFile(tempDir, fileName, content, true)
		require.NoError(t, err)
		
		// Verify file was created with correct content
		filePath := filepath.Join(tempDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(result))
	})

	t.Run("non-safe mode - overwrites existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "overwrite.go"
		originalContent := "// Original content"
		newContent := "// Overwritten content"
		
		// Create initial file
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte(originalContent), 0644)
		require.NoError(t, err)
		
		// Write with non-safe mode - should overwrite
		err = fileio.WriteToFile(tempDir, fileName, newContent, false)
		require.NoError(t, err)
		
		// Verify file was overwritten
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, newContent, string(result))
		assert.NotEqual(t, originalContent, string(result))
	})

	t.Run("empty content writing", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "empty.txt"
		content := ""
		
		err := fileio.WriteToFile(tempDir, fileName, content, false)
		require.NoError(t, err)
		
		// Verify empty file was created
		filePath := filepath.Join(tempDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Empty(t, string(result))
		
		// Verify file exists but is empty
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.Equal(t, int64(0), info.Size())
	})

	t.Run("large content writing", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "large.txt"
		
		// Create large content
		largeContent := make([]byte, 100*1024) // 100KB
		for i := range largeContent {
			largeContent[i] = byte('A' + (i % 26))
		}
		content := string(largeContent)
		
		err := fileio.WriteToFile(tempDir, fileName, content, false)
		require.NoError(t, err)
		
		// Verify large file was written correctly
		filePath := filepath.Join(tempDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(result))
		assert.Len(t, result, 100*1024)
	})

	t.Run("unicode content writing", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "unicode.txt"
		content := "Hello 世界! 🌍 Здравствуй мир! こんにちは世界！"
		
		err := fileio.WriteToFile(tempDir, fileName, content, false)
		require.NoError(t, err)
		
		// Verify unicode content was written correctly
		filePath := filepath.Join(tempDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, string(result))
	})

	t.Run("write to read-only directory fails", func(t *testing.T) {
		// Skip on Windows as permission handling is different
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping permission test on Windows")
		}
		
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly")
		
		// Create directory and make it read-only
		err := os.Mkdir(readOnlyDir, 0755)
		require.NoError(t, err)
		
		err = os.Chmod(readOnlyDir, 0555) // Read + execute only
		require.NoError(t, err)
		
		// Restore permissions after test for cleanup
		t.Cleanup(func() {
			os.Chmod(readOnlyDir, 0755)
		})
		
		fileName := "test.txt"
		content := "test content"
		
		err = fileio.WriteToFile(readOnlyDir, fileName, content, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "permission denied")
	})

	t.Run("creates deeply nested directories", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedDir := filepath.Join(tempDir, "a", "b", "c")

		err := fileio.WriteToFile(nestedDir, "out.go", "content", false)
		require.NoError(t, err)

		result, err := os.ReadFile(filepath.Join(nestedDir, "out.go"))
		require.NoError(t, err)
		assert.Equal(t, "content", string(result))
	})

	t.Run("succeeds when directory already exists", func(t *testing.T) {
		// os.Mkdir returns ErrExist for an existing directory; the fix explicitly ignores
		// that error so repeated writes to the same directory don't fail.
		tempDir := t.TempDir()
		fileName := "repeat.go"
		content := "package main"

		err := fileio.WriteToFile(tempDir, fileName, content, false)
		require.NoError(t, err)

		// Write again to the same (now-existing) directory
		err = fileio.WriteToFile(tempDir, fileName, content, false)
		require.NoError(t, err)
	})

	t.Run("generated code file writing", func(t *testing.T) {
		// Test scenario matching actual usage in code generator
		tempDir := t.TempDir()
		fileName := "server.go"
		generatedContent := `// Package generated by API code generator
package server

import (
	"context"
	"net/http"
)

type Server struct {
	// Server implementation
}

func NewServer() *Server {
	return &Server{}
}
`
		
		err := fileio.WriteToFile(tempDir, fileName, generatedContent, false)
		require.NoError(t, err)
		
		// Verify generated file structure
		filePath := filepath.Join(tempDir, fileName)
		result, err := os.ReadFile(filePath)
		require.NoError(t, err)
		
		content := string(result)
		assert.Contains(t, content, "package server")
		assert.Contains(t, content, "Package generated by API code generator")
		assert.Contains(t, content, "func NewServer()")
	})
}