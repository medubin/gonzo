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

// TestTypeScript_EnumHelpers verifies that isValid and parse helpers are generated
// for every enum, with the correct base type on the parse function.
func TestTypeScript_EnumHelpers(t *testing.T) {
	const api = `
enum Color string {
  RED = "red"
  GREEN = "green"
  BLUE = "blue"
}

enum Priority int32 {
  LOW = 0
  MEDIUM = 1
  HIGH = 2
}

server TestServer { GetFoo GET /foo }
`
	parser := generator.NewParser(api)
	parsed, err := parser.Parse()
	require.NoError(t, err)

	g, err := generator.NewTemplateGenerator("languages/typescript/client/config.yaml")
	require.NoError(t, err)

	results, err := g.Generate(parsed, "client")
	require.NoError(t, err)

	types := results["types.ts"]
	require.NotEmpty(t, types)

	// String enum — values array contains all values, type guard, parse with string param
	assert.Contains(t, types, `const colorValues = [`)
	assert.Contains(t, types, `"red"`)
	assert.Contains(t, types, `"green"`)
	assert.Contains(t, types, `"blue"`)
	assert.Contains(t, types, `] as const`)
	assert.Contains(t, types, `export function isValidColor(v: unknown): v is Color`)
	assert.Contains(t, types, `export function parseColor(v: string): Color`)
	assert.Contains(t, types, `throw new Error`)

	// Numeric enum — parse takes number, not string
	assert.Contains(t, types, `const priorityValues = [`)
	assert.Contains(t, types, `export function isValidPriority(v: unknown): v is Priority`)
	assert.Contains(t, types, `export function parsePriority(v: number): Priority`)

	// isValid uses the values array with the unknown cast pattern
	assert.Contains(t, types, `(colorValues as readonly unknown[]).includes(v)`)
}

// TestTypeScript_ErrorsFileMirrorsGerrors verifies that the generated errors.ts
// contains a class and ErrorCode entry for every exported error type in gerrors.
// This test catches drift between the Go error definitions and the TypeScript output.
func TestTypeScript_ErrorsFileMirrorsGerrors(t *testing.T) {
	parser := generator.NewParser(`server TestServer { GetFoo GET /foo }`)
	parsed, err := parser.Parse()
	require.NoError(t, err)

	g, err := generator.NewTemplateGenerator("languages/typescript/client/config.yaml")
	require.NoError(t, err)

	results, err := g.Generate(parsed, "client")
	require.NoError(t, err)

	errorsTS := results["errors.ts"]
	require.NotEmpty(t, errorsTS)

	// Each of these must appear as both a class definition and a case in the switch.
	// Update this list whenever a new public error type is added to runtime/gerrors.
	expectedErrors := []struct {
		className string
		code      string
	}{
		{"NotFoundError", "not_found"},
		{"InvalidArgumentError", "invalid_argument"},
		{"MissingArgumentError", "missing_argument"},
		{"AlreadyExistsError", "already_exists"},
		{"UnauthenticatedError", "unauthenticated"},
		{"UnimplementedError", "unimplemented"},
		{"InternalError", "internal"},
	}

	for _, e := range expectedErrors {
		assert.Contains(t, errorsTS, "class "+e.className, "errors.ts missing class %s", e.className)
		assert.Contains(t, errorsTS, "'"+e.code+"'", "errors.ts missing error code '%s'", e.code)
	}

	assert.Contains(t, errorsTS, "export type ErrorCode")
	assert.Contains(t, errorsTS, "parseGonzoError")
	assert.Contains(t, errorsTS, "extends GonzoError")
}

// TestTypeScript_ClientUsesParseGonzoError verifies the generated client.ts throws
// typed errors via parseGonzoError rather than a generic Error.
func TestTypeScript_ClientUsesParseGonzoError(t *testing.T) {
	const api = `
type Item { Name string }
server ItemService {
  GetItem GET /items/{id string} returns(Item)
  CreateItem POST /items body(Item) returns(Item)
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

	assert.Contains(t, client, "parseGonzoError", "client should import and use parseGonzoError")
	assert.Contains(t, client, "throw await parseGonzoError(response)", "client should throw typed errors")
	assert.NotContains(t, client, "new Error(", "client should not throw generic Error objects")
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
