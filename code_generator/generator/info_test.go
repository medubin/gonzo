package generator_test

import (
	"strings"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo_AllFieldsParseAndRender(t *testing.T) {
	api, err := generator.NewParser(`
info {
  title "Customer API"
  version "1.4.2"
  description "Public-facing customer endpoints"
  contact "api@example.com"
  license "Apache-2.0"
}

type User { Name string }
server UserService {
  GetUser GET /users/{id int64} returns(User)
}
`).Parse()
	require.NoError(t, err)
	require.NotNil(t, api.Info)
	assert.Equal(t, "Customer API", api.Info.Title)
	assert.Equal(t, "1.4.2", api.Info.Version)
	assert.Equal(t, "Public-facing customer endpoints", api.Info.Description)
	assert.Equal(t, "api@example.com", api.Info.Contact)
	assert.Equal(t, "Apache-2.0", api.Info.License)

	out, err := generator.RenderOpenAPI(api, "fallback-title")
	require.NoError(t, err)
	// info.title takes precedence over the fallback title argument.
	assert.Contains(t, out, "title: Customer API")
	assert.NotContains(t, out, "fallback-title")
	assert.Contains(t, out, "version: 1.4.2")
	assert.Contains(t, out, "description: Public-facing customer endpoints")
	// `@` in contact triggers email mapping.
	assert.Contains(t, out, "contact:\n    email: api@example.com")
	assert.Contains(t, out, "license:\n    name: Apache-2.0")
}

func TestInfo_FallsBackToTitleArgWhenAbsent(t *testing.T) {
	api, err := generator.NewParser(`
type X { N string }
server S { Get GET /x returns(X) }
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "MyAPI")
	require.NoError(t, err)
	assert.Contains(t, out, "title: MyAPI")
	assert.Contains(t, out, "version: 0.0.0")
}

func TestInfo_ContactWithoutEmailRendersAsName(t *testing.T) {
	api, err := generator.NewParser(`
info {
  contact "Platform Team"
}
type X { N string }
server S { Get GET /x returns(X) }
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "X")
	require.NoError(t, err)
	assert.Contains(t, out, "contact:\n    name: Platform Team")
	assert.NotContains(t, out, "email:")
}

func TestInfo_DuplicateBlockErrors(t *testing.T) {
	_, err := generator.NewParser(`
info { version "1" }
info { version "2" }
type X { N string }
server S { Get GET /x returns(X) }
`).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "info block already defined")
}

func TestInfo_DuplicateFieldErrors(t *testing.T) {
	_, err := generator.NewParser(`
info {
  version "1"
  version "2"
}
type X { N string }
server S { Get GET /x returns(X) }
`).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `info: field "version" already set`)
}

func TestInfo_UnknownFieldErrors(t *testing.T) {
	_, err := generator.NewParser(`
info { author "alice" }
type X { N string }
server S { Get GET /x returns(X) }
`).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown field "author"`)
}

func TestOpenAPI_TagsPerServer(t *testing.T) {
	api, err := generator.NewParser(`
type User { Name string }
type Photo { Url string }

server UserService {
  GetUser GET /users/{id int64} returns(User)
}
server MediaService {
  GetPhoto GET /photos/{id int64} returns(Photo)
}
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "Multi")
	require.NoError(t, err)

	// Top-level tags block lists every server.
	assert.Contains(t, out, "tags:\n  - name: UserService\n  - name: MediaService\n")

	// Each operation references its server.
	getUserIdx := strings.Index(out, "operationId: GetUser")
	getPhotoIdx := strings.Index(out, "operationId: GetPhoto")
	require.True(t, getUserIdx > 0 && getPhotoIdx > 0)
	// The `tags:` for the operation appears before the operationId line.
	assert.Contains(t, out[:getUserIdx], "tags:\n        - UserService")
	assert.Contains(t, out[:getPhotoIdx], "tags:\n        - MediaService")
}
