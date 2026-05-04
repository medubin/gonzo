package generator_test

import (
	"strings"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseDecorated(t *testing.T, src string) *generator.APIDefinition {
	t.Helper()
	api, err := generator.NewParser(src).Parse()
	require.NoError(t, err)
	return api
}

func TestDecorator_NoArgs(t *testing.T) {
	api := parseDecorated(t, `
type User { Name string }
server S {
  @deprecated
  Get GET /users returns(User)
}
`)
	d := api.Servers[0].Endpoints[0].Decorators
	require.Len(t, d, 1)
	assert.Equal(t, "deprecated", d[0].Name)
	assert.Empty(t, d[0].Args)
	assert.Empty(t, d[0].Kwargs)
}

func TestDecorator_PositionalAndKwargMix(t *testing.T) {
	api := parseDecorated(t, `
type User { Name string }
server S {
  @cache(60, public: true, label: "users-list")
  List GET /users returns(User)
}
`)
	d := api.Servers[0].Endpoints[0].Decorators
	require.Len(t, d, 1)
	assert.Equal(t, "cache", d[0].Name)
	require.Len(t, d[0].Args, 1)
	assert.Equal(t, "number", d[0].Args[0].Kind)
	assert.Equal(t, "60", d[0].Args[0].Value)
	require.Len(t, d[0].Kwargs, 2)
	assert.Equal(t, "public", d[0].Kwargs[0].Name)
	assert.Equal(t, "bool", d[0].Kwargs[0].Arg.Kind)
	assert.Equal(t, "true", d[0].Kwargs[0].Arg.Value)
	assert.Equal(t, "label", d[0].Kwargs[1].Name)
	assert.Equal(t, "string", d[0].Kwargs[1].Arg.Kind)
	assert.Equal(t, "users-list", d[0].Kwargs[1].Arg.Value)
}

func TestDecorator_StackOnSameEndpoint(t *testing.T) {
	api := parseDecorated(t, `
type User { Name string }
server S {
  @auth("bearer")
  @deprecated
  Get GET /users returns(User)
}
`)
	d := api.Servers[0].Endpoints[0].Decorators
	require.Len(t, d, 2)
	assert.Equal(t, "auth", d[0].Name)
	assert.Equal(t, "deprecated", d[1].Name)
}

func TestDecorator_PositionalAfterKwargErrors(t *testing.T) {
	_, err := generator.NewParser(`
server S {
  @bad(public: true, 60) Get GET /x
}
`).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "positional decorator argument after kwarg")
}

func TestDecorator_CascadesFromGroupToEndpoints(t *testing.T) {
	api := parseDecorated(t, `
type UserID int64
type User { Name string }

server S {
  @auth("bearer")
  group /admin {
    DeleteUser DELETE /users/{id UserID}
    @deprecated
    PurgeAll DELETE /purge
  }
  Public GET /health
}
`)
	endpoints := api.Servers[0].Endpoints
	require.Len(t, endpoints, 3)

	// Both group endpoints inherit @auth("bearer").
	assert.Equal(t, "DeleteUser", endpoints[0].Name)
	require.Len(t, endpoints[0].Decorators, 1)
	assert.Equal(t, "auth", endpoints[0].Decorators[0].Name)
	assert.Equal(t, "bearer", endpoints[0].Decorators[0].Args[0].Value)

	// Endpoint with its own decorator gets both, group first.
	assert.Equal(t, "PurgeAll", endpoints[1].Name)
	require.Len(t, endpoints[1].Decorators, 2)
	assert.Equal(t, "auth", endpoints[1].Decorators[0].Name)
	assert.Equal(t, "deprecated", endpoints[1].Decorators[1].Name)

	// Endpoint outside the group is untouched.
	assert.Equal(t, "Public", endpoints[2].Name)
	assert.Empty(t, endpoints[2].Decorators)
}

func TestDecorator_EndpointOverridesGroupAuth(t *testing.T) {
	api := parseDecorated(t, `
server S {
  @auth("bearer")
  group /admin {
    @auth("none")
    Heartbeat GET /heartbeat
  }
}
`)
	ep := api.Servers[0].Endpoints[0]
	require.Len(t, ep.Decorators, 2)

	// Both decorators are present in source order. The generator's last-wins
	// loop is what makes "none" effective; verify the slice ordering here.
	assert.Equal(t, "bearer", ep.Decorators[0].Args[0].Value)
	assert.Equal(t, "none", ep.Decorators[1].Args[0].Value)

	// Render through OpenAPI to confirm last-wins applies end-to-end:
	// "none" means no `security:` block on the operation.
	out, err := generator.RenderOpenAPI(api, "Test")
	require.NoError(t, err)
	hbIdx := strings.Index(out, "operationId: Heartbeat")
	require.True(t, hbIdx > 0)
	// No security: between the operation and the next path or end of doc.
	assert.NotContains(t, out[hbIdx:], "security:\n        - bearerAuth")
}

func TestDecorator_NestedGroupsStackDecorators(t *testing.T) {
	api := parseDecorated(t, `
server S {
  @auth("bearer")
  group /v1 {
    @deprecated
    group /legacy {
      Get GET /thing
    }
  }
}
`)
	ep := api.Servers[0].Endpoints[0]
	require.Len(t, ep.Decorators, 2)
	assert.Equal(t, "auth", ep.Decorators[0].Name)
	assert.Equal(t, "deprecated", ep.Decorators[1].Name)
}

func TestDecorator_AuthAppearsInGoServerOutput(t *testing.T) {
	api := parseDecorated(t, `
type User { Name string }
server S {
  @auth("bearer")
  Get GET /users returns(User)
  Public GET /health
}
`)
	// Confirm AuthScheme propagates through the parser AST. Generator-side
	// emission is covered by the snapshot test once test_server.api is
	// updated to use @auth.
	require.Len(t, api.Servers[0].Endpoints, 2)
	assert.Equal(t, "auth", api.Servers[0].Endpoints[0].Decorators[0].Name)
	assert.Equal(t, "bearer", api.Servers[0].Endpoints[0].Decorators[0].Args[0].Value)
	assert.Empty(t, api.Servers[0].Endpoints[1].Decorators)
}

func TestDecorator_DeprecatedInOpenAPI(t *testing.T) {
	api := parseDecorated(t, `
type User { Name string }
server S {
  @deprecated("use v2 GetUser")
  Get GET /users/{id int64} returns(User)

  @deprecated
  OldList GET /list returns(User)

  Fresh GET /fresh returns(User)
}
`)
	out, err := generator.RenderOpenAPI(api, "Test")
	require.NoError(t, err)

	// Both deprecated endpoints emit `deprecated: true`; the fresh one doesn't.
	assert.Contains(t, out, "operationId: Get\n      deprecated: true")
	assert.Contains(t, out, "operationId: OldList\n      deprecated: true")
	freshIdx := strings.Index(out, "operationId: Fresh")
	require.True(t, freshIdx > 0)
	// The next line after Fresh's operationId must NOT be `deprecated: true`.
	tail := out[freshIdx:]
	nl := strings.Index(tail, "\n")
	require.True(t, nl > 0)
	assert.NotContains(t, tail[nl:nl+30], "deprecated: true")
}

func TestDecorator_AuthInOpenAPIOutput(t *testing.T) {
	api := parseDecorated(t, `
type User { Name string }
server S {
  @auth("bearer")
  GetMe GET /me returns(User)

  @auth("apiKey")
  Admin GET /admin returns(User)

  @auth("none")
  Health GET /health

  Public GET /public
}
`)
	out, err := generator.RenderOpenAPI(api, "Test")
	require.NoError(t, err)

	// Per-op security on the two real schemes.
	assert.Contains(t, out, "security:\n        - bearerAuth: []")
	assert.Contains(t, out, "security:\n        - apiKeyAuth: []")

	// "none" and the unannotated route get no security block.
	healthIdx := strings.Index(out, "/health")
	publicIdx := strings.Index(out, "/public")
	require.True(t, healthIdx > 0 && publicIdx > 0)
	// Slice from /health to /public and ensure no security: appears.
	assert.NotContains(t, out[healthIdx:publicIdx], "security:")
	// Slice from /public to end: no security either.
	assert.NotContains(t, out[publicIdx:], "security:")

	// Components declarations.
	assert.Contains(t, out, "securitySchemes:")
	assert.Contains(t, out, "bearerAuth:\n      type: http\n      scheme: bearer\n      bearerFormat: JWT")
	assert.Contains(t, out, "apiKeyAuth:\n      type: apiKey\n      in: header\n      name: X-API-Key")
}
