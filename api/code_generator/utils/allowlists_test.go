package utils_test

import (
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/medubin/gonzo/api/code_generator/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetLanguageStackConfig(t *testing.T) {
	t.Run("supported Go server configuration", func(t *testing.T) {
		result := utils.GetLanguageStackConfig("go", "server")
		expected := "api/code_generator/generator/languages/go/server/config.yaml"
		assert.Equal(t, expected, result)
	})

	t.Run("supported TypeScript client configuration", func(t *testing.T) {
		result := utils.GetLanguageStackConfig("typescript", "client")
		expected := "api/code_generator/generator/languages/typescript/client/config.yaml"
		assert.Equal(t, expected, result)
	})

	t.Run("unsupported language returns empty string", func(t *testing.T) {
		testCases := []struct {
			name     string
			language string
			stack    string
		}{
			{"python server", "python", "server"},
			{"java server", "java", "server"},
			{"rust server", "rust", "server"},
			{"javascript client", "javascript", "client"},
			{"dart client", "dart", "client"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := utils.GetLanguageStackConfig(tc.language, tc.stack)
				assert.Empty(t, result, "Expected empty string for unsupported language %s with stack %s", tc.language, tc.stack)
			})
		}
	})

	t.Run("unsupported stack returns empty string", func(t *testing.T) {
		testCases := []struct {
			name     string
			language string
			stack    string
		}{
			{"go mobile", "go", "mobile"},
			{"go desktop", "go", "desktop"},
			{"typescript server", "typescript", "server"},
			{"typescript mobile", "typescript", "mobile"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := utils.GetLanguageStackConfig(tc.language, tc.stack)
				assert.Empty(t, result, "Expected empty string for language %s with unsupported stack %s", tc.language, tc.stack)
			})
		}
	})

	t.Run("empty inputs return empty string", func(t *testing.T) {
		testCases := []struct {
			name     string
			language string
			stack    string
		}{
			{"empty language", "", "server"},
			{"empty stack", "go", ""},
			{"both empty", "", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := utils.GetLanguageStackConfig(tc.language, tc.stack)
				assert.Empty(t, result, "Expected empty string for empty inputs")
			})
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		testCases := []struct {
			name     string
			language string
			stack    string
		}{
			{"uppercase Go", "GO", "server"},
			{"uppercase TypeScript", "TYPESCRIPT", "client"},
			{"mixed case Go", "Go", "server"},
			{"mixed case server", "go", "Server"},
			{"mixed case client", "typescript", "Client"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := utils.GetLanguageStackConfig(tc.language, tc.stack)
				assert.Empty(t, result, "Expected empty string for case variations - function should be case sensitive")
			})
		}
	})

	t.Run("whitespace handling", func(t *testing.T) {
		testCases := []struct {
			name     string
			language string
			stack    string
		}{
			{"leading space language", " go", "server"},
			{"trailing space language", "go ", "server"},
			{"leading space stack", "go", " server"},
			{"trailing space stack", "go", "server "},
			{"language with spaces", "go lang", "server"},
			{"stack with spaces", "go", "server stack"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := utils.GetLanguageStackConfig(tc.language, tc.stack)
				assert.Empty(t, result, "Expected empty string for inputs with whitespace")
			})
		}
	})

	t.Run("all supported combinations snapshot", func(t *testing.T) {
		// Create a map of all supported combinations for snapshot testing
		supportedCombinations := map[string]string{
			"go+server":         utils.GetLanguageStackConfig("go", "server"),
			"typescript+client": utils.GetLanguageStackConfig("typescript", "client"),
		}

		// This helps document what combinations are currently supported
		snaps.MatchSnapshot(t, supportedCombinations)
	})

	t.Run("configuration path format validation", func(t *testing.T) {
		// Test that returned paths follow expected format
		testCases := []struct {
			language string
			stack    string
		}{
			{"go", "server"},
			{"typescript", "client"},
		}

		for _, tc := range testCases {
			result := utils.GetLanguageStackConfig(tc.language, tc.stack)
			if result != "" {
				// Verify path format: api/code_generator/generator/languages/{language}/{stack}/config.yaml
				assert.Contains(t, result, "api/code_generator/generator/languages/")
				assert.Contains(t, result, tc.language)
				assert.Contains(t, result, tc.stack)
				assert.True(t, strings.HasSuffix(result, "config.yaml"))
			}
		}
	})

	t.Run("exhaustive combination testing", func(t *testing.T) {
		// Test all possible combinations systematically
		languages := []string{"go", "typescript", "python", "java", "rust", "javascript", "c#", "kotlin"}
		stacks := []string{"server", "client", "mobile", "desktop", "web", "api"}

		supportedCount := 0
		unsupportedCount := 0

		for _, lang := range languages {
			for _, stack := range stacks {
				result := utils.GetLanguageStackConfig(lang, stack)
				if result != "" {
					supportedCount++
					// Verify supported combinations return valid paths
					assert.Contains(t, result, lang)
					assert.Contains(t, result, stack)
				} else {
					unsupportedCount++
				}
			}
		}

		// Document current support level
		t.Logf("Supported combinations: %d", supportedCount)
		t.Logf("Unsupported combinations: %d", unsupportedCount)
		
		// Verify we have exactly the expected supported combinations
		assert.Equal(t, 2, supportedCount, "Expected exactly 2 supported combinations (go+server, typescript+client)")
		assert.Equal(t, 46, unsupportedCount, "Expected 46 unsupported combinations out of 48 total")
	})
}