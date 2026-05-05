package generator_test

import (
	"strings"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// renderForTest parses src and returns the OpenAPI document. Helper to keep
// the assertions below uncluttered.
func renderForTest(t *testing.T, src string) string {
	t.Helper()
	api, err := generator.NewParser(src).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)
	return out
}

func TestResponse_AdditionalErrorCodesPreserveImplicit200(t *testing.T) {
	// Only non-2xx declared → implicit 200 from `returns` stays.
	out := renderForTest(t, `
type User { Name string }
type NotFoundError { msg string }
server S {
  @response(404, NotFoundError)
  Get GET /users/{id int64} returns(User)
}
`)
	assert.Contains(t, out, "'200':")
	assert.Contains(t, out, "'404':")
	assert.Contains(t, out, "description: Not Found")
	assert.Contains(t, out, "$ref: '#/components/schemas/NotFoundError'")
}

func TestResponse_DeclaredSuccessSuppresses200(t *testing.T) {
	// @response(201, ...) suppresses the implicit 200.
	out := renderForTest(t, `
type User { Name string }
server S {
  @response(201, User)
  Create POST /users body(User) returns(User)
}
`)
	assert.Contains(t, out, "'201':")
	assert.Contains(t, out, "description: Created")
	createIdx := strings.Index(out, "operationId: Create")
	defaultIdx := strings.Index(out[createIdx:], "default:")
	require.True(t, defaultIdx > 0)
	// Between the operationId and `default:`, no 200 entry should appear.
	assert.NotContains(t, out[createIdx:createIdx+defaultIdx], "'200':")
}

func TestResponse_BodylessCodes(t *testing.T) {
	// `@response(204)` and `@response(304)` with no type emit no `content:`.
	out := renderForTest(t, `
type User { Name string }
server S {
  @response(204)
  @response(304)
  Get GET /users/{id int64} returns(User)
}
`)
	// 204 declared (a 2xx) → implicit 200 suppressed.
	assert.NotContains(t, out, "'200':")
	assert.Contains(t, out, "'204':")
	assert.Contains(t, out, "'304':")
	// No content blocks for these — body type was not provided. Slice the
	// portion of the doc between '204' and the next response key.
	idx204 := strings.Index(out, "'204':")
	idx304 := strings.Index(out, "'304':")
	idxDefault := strings.Index(out, "default:")
	require.True(t, idx204 > 0 && idx304 > 0 && idxDefault > 0)
	assert.NotContains(t, out[idx204:idx304], "content:")
	assert.NotContains(t, out[idx304:idxDefault], "content:")
}

func TestResponse_ExplicitDescriptionOverride(t *testing.T) {
	out := renderForTest(t, `
type X { N string }
server S {
  @response(409, X, description: "Username taken")
  Create POST /users body(X) returns(X)
}
`)
	assert.Contains(t, out, "description: Username taken")
	// Default reason phrase ("Conflict") should NOT appear since user overrode it.
	conflictIdx := strings.Index(out, "'409':")
	require.True(t, conflictIdx > 0)
	defaultIdx := strings.Index(out[conflictIdx:], "default:")
	assert.NotContains(t, out[conflictIdx:conflictIdx+defaultIdx], "description: Conflict")
}

func TestResponse_BareIdentifierTypeArgWorks(t *testing.T) {
	// @response(201, User) — User is a bare identifier, not a quoted string.
	// The parser/extractor should accept it as a type reference.
	out := renderForTest(t, `
type User { Name string }
server S {
  @response(201, User)
  Create POST /users body(User) returns(User)
}
`)
	assert.Contains(t, out, "'201':")
	assert.Contains(t, out, "$ref: '#/components/schemas/User'")
}
