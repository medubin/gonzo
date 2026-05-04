package generator_test

import (
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parse(t *testing.T, src string) *generator.APIDefinition {
	t.Helper()
	api, err := generator.NewParser(src).Parse()
	require.NoError(t, err)
	return api
}

func TestGroup_FlattensPrefixAndParams(t *testing.T) {
	api := parse(t, `
type UserID int64
type User { ID UserID }

server S {
  group /users/{id UserID} {
    Get GET returns(User)
    Patch PATCH /profile body(User) returns(User)
  }
}
`)

	require.Len(t, api.Servers, 1)
	endpoints := api.Servers[0].Endpoints
	require.Len(t, endpoints, 2)

	assert.Equal(t, "Get", endpoints[0].Name)
	assert.Equal(t, "/users/{id}", endpoints[0].Path)
	require.Len(t, endpoints[0].PathParams, 1)
	assert.Equal(t, "id", endpoints[0].PathParams[0].Name)
	assert.Equal(t, "UserID", endpoints[0].PathParams[0].Type)

	assert.Equal(t, "Patch", endpoints[1].Name)
	assert.Equal(t, "/users/{id}/profile", endpoints[1].Path)
	require.Len(t, endpoints[1].PathParams, 1)
}

func TestGroup_NestedGroupsStack(t *testing.T) {
	api := parse(t, `
type UserID int64
type NotificationID int32
type Notification { ID NotificationID }

server S {
  group /users/{id UserID} {
    group /notifications {
      List GET returns(Notification)
      MarkRead PUT /{nid NotificationID}
    }
  }
}
`)

	endpoints := api.Servers[0].Endpoints
	require.Len(t, endpoints, 2)
	assert.Equal(t, "/users/{id}/notifications", endpoints[0].Path)
	assert.Equal(t, []generator.ParamDef{{Name: "id", Type: "UserID"}}, endpoints[0].PathParams)
	assert.Equal(t, "/users/{id}/notifications/{nid}", endpoints[1].Path)
	assert.Equal(t, []generator.ParamDef{
		{Name: "id", Type: "UserID"},
		{Name: "nid", Type: "NotificationID"},
	}, endpoints[1].PathParams)
}

func TestGroup_MixWithUngroupedEndpoints(t *testing.T) {
	api := parse(t, `
type UserID int64
type User { ID UserID }
type CreateReq { Name string }

server S {
  Create POST /users body(CreateReq) returns(User)
  group /users/{id UserID} {
    Get GET returns(User)
  }
  Health GET /health
}
`)
	endpoints := api.Servers[0].Endpoints
	require.Len(t, endpoints, 3)
	assert.Equal(t, "/users", endpoints[0].Path)
	assert.Equal(t, "/users/{id}", endpoints[1].Path)
	assert.Equal(t, "/health", endpoints[2].Path)
}

func TestGroup_DuplicateParamAcrossLevelsErrors(t *testing.T) {
	_, err := generator.NewParser(`
type UserID int64

server S {
  group /users/{id UserID} {
    Get GET /{id UserID}
  }
}
`).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate path parameter")
}

func TestGroup_EndpointWithoutPathOutsideGroupErrors(t *testing.T) {
	_, err := generator.NewParser(`
type User { Name string }

server S {
  Get GET returns(User)
}
`).Parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected path starting with '/'")
}
