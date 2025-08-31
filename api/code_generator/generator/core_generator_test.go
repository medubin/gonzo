package generator_test

import (
	"sort"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/medubin/gonzo/api/code_generator/fileio"
	"github.com/medubin/gonzo/api/code_generator/generator"
	"github.com/stretchr/testify/require"
)

func TestCoreGenerate_Go_Snapshot(t *testing.T) {
	// Parse the test API
	input, err := fileio.ParseFile("test_data/test_server.api")
	require.NoError(t, err)

	parser := generator.NewParser(string(input))
	api, err := parser.Parse()
	require.NoError(t, err)

	// Generate Go server code
	g, err := generator.NewTemplateGenerator("languages/go/server/config.yaml")
	require.NoError(t, err)
	
	results, err := g.Generate(api, "server")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Sort filenames for deterministic testing
	filenames := make([]string, 0, len(results))
	for filename := range results {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// Create snapshot for each generated file
	for _, filename := range filenames {
		content := results[filename]
		// Use a descriptive test name that includes the filename
		t.Run(filename, func(t *testing.T) {
			snaps.MatchStandaloneSnapshot(t, content)
		})
	}
}

func TestCoreGenerate_TypeScript_Snapshot(t *testing.T) {
	// Parse the test API
	input, err := fileio.ParseFile("test_data/test_server.api")
	require.NoError(t, err)

	parser := generator.NewParser(string(input))
	api, err := parser.Parse()
	require.NoError(t, err)

	// Generate TypeScript client code
	g, err := generator.NewTemplateGenerator("languages/typescript/client/config.yaml")
	require.NoError(t, err)
	
	results, err := g.Generate(api, "client")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Sort filenames for deterministic testing
	filenames := make([]string, 0, len(results))
	for filename := range results {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	// Create snapshot for each generated file
	for _, filename := range filenames {
		content := results[filename]
		// Use a descriptive test name that includes the filename
		t.Run(filename, func(t *testing.T) {
			snaps.MatchStandaloneSnapshot(t, content)
		})
	}
}
