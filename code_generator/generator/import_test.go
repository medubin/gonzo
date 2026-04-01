package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestImport_MergesTypesAndEnums(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "common.api", `
type SharedID int64

enum Color string {
  RED = "red"
  BLUE = "blue"
}
`)

	mainPath := writeFile(t, dir, "main.api", `
import "common.api"

type LocalType string

server MyServer {
  GetFoo GET /foo
}
`)

	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	p := generator.NewParser(string(data), mainPath)
	api, err := p.Parse()
	require.NoError(t, err)

	// Types from both files
	typeNames := make(map[string]bool)
	for _, tt := range api.Types {
		typeNames[tt.Name] = true
	}
	assert.True(t, typeNames["SharedID"], "SharedID from common.api should be present")
	assert.True(t, typeNames["LocalType"], "LocalType from main.api should be present")

	// Enum from common.api
	require.Len(t, api.Enums, 1)
	assert.Equal(t, "Color", api.Enums[0].Name)

	// Server from main.api
	require.Len(t, api.Servers, 1)
	assert.Equal(t, "MyServer", api.Servers[0].Name)
}

func TestImport_TransitiveImports(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "base.api", `type BaseType string`)
	writeFile(t, dir, "middle.api", `
import "base.api"
type MiddleType int32
`)
	mainPath := writeFile(t, dir, "main.api", `
import "middle.api"
type TopType bool
`)

	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	p := generator.NewParser(string(data), mainPath)
	api, err := p.Parse()
	require.NoError(t, err)

	typeNames := make(map[string]bool)
	for _, tt := range api.Types {
		typeNames[tt.Name] = true
	}
	assert.True(t, typeNames["BaseType"])
	assert.True(t, typeNames["MiddleType"])
	assert.True(t, typeNames["TopType"])
}

func TestImport_CircularImportSkipped(t *testing.T) {
	dir := t.TempDir()

	// a.api imports b.api, b.api imports a.api
	writeFile(t, dir, "a.api", `
import "b.api"
type TypeA string
`)
	writeFile(t, dir, "b.api", `
import "a.api"
type TypeB string
`)

	mainPath := filepath.Join(dir, "a.api")
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	p := generator.NewParser(string(data), mainPath)
	api, err := p.Parse()
	require.NoError(t, err)

	typeNames := make(map[string]bool)
	for _, tt := range api.Types {
		typeNames[tt.Name] = true
	}
	assert.True(t, typeNames["TypeA"])
	assert.True(t, typeNames["TypeB"])
}

func TestImport_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	mainPath := writeFile(t, dir, "main.api", `import "nonexistent.api"`)

	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	p := generator.NewParser(string(data), mainPath)
	_, err = p.Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent.api")
}

func TestImport_MissingStringPath(t *testing.T) {
	p := generator.NewParser(`import type Foo string`, "/tmp")
	_, err := p.Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected string path after import")
}

func TestImport_NoBadSideEffectsWithoutBaseDir(t *testing.T) {
	// When no baseDir is provided and there are no imports, parsing still works.
	p := generator.NewParser(`type Foo string`)
	api, err := p.Parse()
	require.NoError(t, err)
	require.Len(t, api.Types, 1)
	assert.Equal(t, "Foo", api.Types[0].Name)
}

func TestImport_ConflictingTypeFromImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "common.api", `type User string`)
	mainPath := writeFile(t, dir, "main.api", `
import "common.api"
type User string
`)
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	_, err = generator.NewParser(string(data), mainPath).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `type "User" is already defined`)
}

func TestImport_ConflictingEnumFromImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "common.api", `enum Color string { RED = "red" }`)
	mainPath := writeFile(t, dir, "main.api", `
import "common.api"
enum Color string { BLUE = "blue" }
`)
	data, err := os.ReadFile(mainPath)
	require.NoError(t, err)

	_, err = generator.NewParser(string(data), mainPath).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `enum "Color" is already defined`)
}

func TestImport_DuplicateTypeInSameFile(t *testing.T) {
	p := generator.NewParser(`
type User string
type User int32
`)
	_, err := p.Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `type "User" is already defined`)
}

func TestImport_DuplicateServerInSameFile(t *testing.T) {
	p := generator.NewParser(`
server MyService {
  GetFoo GET /foo
}
server MyService {
  GetBar GET /bar
}
`)
	_, err := p.Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `server "MyService" is already defined`)
}
