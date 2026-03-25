package generator_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/medubin/gonzo/code_generator/fileio"
	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTypeScript_ImportOnlyDirectlyReferencedTypes guards against a regression
// where getUsedTypes recurses into struct fields, causing types that are only
// used as fields of other types (never directly by an endpoint) to appear in
// the generated import statement.
func TestTypeScript_ImportOnlyDirectlyReferencedTypes(t *testing.T) {
	const api = `
type NestedType {
  Value string
}

type ParentType {
  Nested NestedType
}

server TestServer {
  GetParent GET /parent returns(ParentType)
}
`
	parser := generator.NewParser(api)
	parsed, err := parser.Parse()
	require.NoError(t, err)

	g, err := generator.NewTemplateGenerator("languages/typescript/client/config.yaml")
	require.NoError(t, err)

	results, err := g.Generate(parsed, "client")
	require.NoError(t, err)

	client := results["client.ts"]
	require.NotEmpty(t, client)

	// ParentType is the direct return type — it must be imported.
	assert.Contains(t, client, "ParentType")

	// NestedType is only a field of ParentType, never directly referenced by
	// an endpoint. It must NOT appear in the import statement.
	importLine := ""
	for _, line := range strings.Split(client, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "import type") {
			importLine = line
			break
		}
	}
	assert.NotContains(t, importLine, "NestedType")
}

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
