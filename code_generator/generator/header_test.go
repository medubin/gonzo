package generator_test

import (
	"strings"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeader_EmittedInOpenAPI(t *testing.T) {
	api, err := generator.NewParser(`
type User { Name string }
server S {
  @header("X-Tenant-Id", required: true, description: "Tenant identifier")
  @header("Idempotency-Key")
  Create POST /users body(User) returns(User)
}
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)

	// Required tenant header.
	assert.Contains(t, out, "- name: X-Tenant-Id\n          in: header\n          required: true\n          description: Tenant identifier\n")
	// Optional idempotency header (no required, no description).
	assert.Contains(t, out, "- name: Idempotency-Key\n          in: header\n          schema:\n")
}

func TestHeader_MiddlewareReachableViaRouteInfo(t *testing.T) {
	// @header decorators land on RouteInfo.Decorators just like any other
	// decorator, so middleware can iterate and enforce them. This test
	// confirms the AST round-trip; the OpenAPI emission is covered above.
	api, err := generator.NewParser(`
type User { Name string }
server S {
  @header("X-Tenant-Id", required: true)
  Get GET /users returns(User)
}
`).Parse()
	require.NoError(t, err)

	ep := api.Servers[0].Endpoints[0]
	require.Len(t, ep.Decorators, 1)
	assert.Equal(t, "header", ep.Decorators[0].Name)
	require.Len(t, ep.Decorators[0].Args, 1)
	assert.Equal(t, "X-Tenant-Id", ep.Decorators[0].Args[0].Value)
}

func TestHeader_NameFromKwargAlsoWorks(t *testing.T) {
	api, err := generator.NewParser(`
type X { N string }
server S {
  @header(name: "X-Trace-Id")
  Get GET /x returns(X)
}
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)
	assert.Contains(t, out, "- name: X-Trace-Id\n          in: header\n")
}

func TestHeader_MissingNameIsIgnored(t *testing.T) {
	// A `@header` with no name is meaningless; we silently skip it rather
	// than emitting a malformed parameter entry. Parser/extractor stays
	// permissive because the open-vocabulary decorator surface means
	// future args (description-only docs?) shouldn't error.
	api, err := generator.NewParser(`
type X { N string }
server S {
  @header(description: "no name supplied")
  Get GET /x returns(X)
}
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)
	// No header parameter should be emitted.
	getIdx := strings.Index(out, "operationId: Get")
	require.True(t, getIdx > 0)
	// Walk forward to the next operation/path or end and ensure no header param.
	tail := out[getIdx:]
	assert.NotContains(t, tail[:300], "in: header")
}
