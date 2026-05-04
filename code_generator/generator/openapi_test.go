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

func renderOpenAPIFromSrc(t *testing.T, src string) string {
	t.Helper()
	api, err := generator.NewParser(src).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "Test")
	require.NoError(t, err)
	return out
}

func TestOpenAPI_HeaderAndInfo(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `server S { Foo GET /foo }`)
	assert.True(t, strings.HasPrefix(out, "openapi: 3.1.0\n"), "must start with openapi version")
	assert.Contains(t, out, "title: Test")
	assert.Contains(t, out, "version: 0.0.0")
}

func TestOpenAPI_PrimitivePathParam(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `server S { GetThing GET /things/{id int64} }`)
	assert.Contains(t, out, "/things/{id}:")
	assert.Contains(t, out, "- name: id\n          in: path\n          required: true\n          schema:\n            type: integer\n            format: int64")
}

func TestOpenAPI_QueryParamsFlattenStruct(t *testing.T) {
	src := `type ListParams {
  required Page int32
  Limit int32
}
server S { ListThings GET /things parameters(ListParams) }`
	out := renderOpenAPIFromSrc(t, src)
	assert.Contains(t, out, "- name: Page\n          in: query\n          required: true\n          schema:\n            type: integer\n            format: int32")
	assert.Contains(t, out, "- name: Limit\n          in: query\n          schema:\n            type: integer")
	// Optional Limit should NOT have required: true.
	limitBlock := out[strings.Index(out, "- name: Limit"):]
	limitBlock = limitBlock[:strings.Index(limitBlock, "schema:")]
	assert.NotContains(t, limitBlock, "required: true")
}

func TestOpenAPI_RequestBodyJSON(t *testing.T) {
	src := `type Pet { Name string }
server S { CreatePet POST /pets body(Pet) returns(Pet) }`
	out := renderOpenAPIFromSrc(t, src)
	assert.Contains(t, out, "requestBody:\n        required: true\n        content:\n          application/json:\n            schema:\n              $ref: '#/components/schemas/Pet'")
}

func TestOpenAPI_RequestBodyMultipart(t *testing.T) {
	src := `type UploadReq {
  required Image file
  Caption string
}
type UploadRes { Url string }
server S { Upload POST /upload body(UploadReq) returns(UploadRes) }`
	out := renderOpenAPIFromSrc(t, src)
	// Bound the slice from "requestBody:" up to "responses:" so we don't catch
	// the application/json content-type used by the success response body.
	rb := out[strings.Index(out, "requestBody:"):]
	rb = rb[:strings.Index(rb, "responses:")]
	assert.Contains(t, rb, "multipart/form-data:")
	assert.NotContains(t, rb, "application/json:")
	// File field renders as binary string in the schema body
	assert.Contains(t, out, "Image:\n          type: string\n          format: binary")
}

func TestOpenAPI_NoReturnsProduces204(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `server S { Ping HEAD /ping }`)
	assert.Contains(t, out, "responses:\n        '204':\n          description: No Content")
	assert.NotContains(t, strings.SplitAfter(out, "operationId: Ping")[1], "'200':")
}

func TestOpenAPI_AllHTTPMethodsSupported(t *testing.T) {
	src := `server S {
  G GET /a
  P POST /a
  U PUT /a
  D DELETE /a
  C PATCH /a
  H HEAD /a
  O OPTIONS /a
}`
	out := renderOpenAPIFromSrc(t, src)
	for _, m := range []string{"get:", "post:", "put:", "delete:", "patch:", "head:", "options:"} {
		assert.Contains(t, out, "\n    "+m, "method %s should appear under /a", m)
	}
}

func TestOpenAPI_ComponentsForEveryType(t *testing.T) {
	src := `type UserID int64
type User { required ID UserID }
type UserList repeated(User)
type UsersByID map(UserID: User)
enum Role string {
  ADMIN = "admin"
  USER = "user"
}
server S { Foo GET /foo }`
	out := renderOpenAPIFromSrc(t, src)
	for _, schema := range []string{"UserID:", "User:", "UserList:", "UsersByID:", "Role:", "GonzoError:"} {
		assert.Contains(t, out, "    "+schema, "expected schema %s in components", schema)
	}
	// Alphabetical order — Role before User.
	roleIdx := strings.Index(out, "    Role:\n")
	userIdx := strings.Index(out, "    User:\n")
	assert.True(t, roleIdx > 0 && userIdx > roleIdx, "schemas should be sorted alphabetically")
}

func TestOpenAPI_AliasOfPrimitiveInlinesPrimitive(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `type UserID int64
server S { Foo GET /foo }`)
	// The UserID schema body should be the int64 primitive, not a $ref to itself.
	assert.Contains(t, out, "    UserID:\n      type: integer\n      format: int64")
}

func TestOpenAPI_AliasOfNamedTypeUsesRef(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `type ID int64
type UserID ID
server S { Foo GET /foo }`)
	assert.Contains(t, out, "    UserID:\n      $ref: '#/components/schemas/ID'")
}

func TestOpenAPI_StructRequiredFields(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `type T {
  required A string
  B string
  required C string
}
server S { Foo GET /foo }`)
	// Required block lists only A and C, in source order, indented under the T schema.
	assert.Contains(t, out, "      required:\n        - A\n        - C\n")
	// B is optional, so it must not appear in the required list.
	assert.NotContains(t, out, "        - B\n")
}

func TestOpenAPI_NestedRepeatedAndMap(t *testing.T) {
	src := `type Nested map(string: repeated(int32))
server S { Foo GET /foo returns(Nested) }`
	out := renderOpenAPIFromSrc(t, src)
	// Outer map → object/additionalProperties; inner repeated → array/items
	assert.Contains(t, out, "    Nested:\n      type: object\n      additionalProperties:\n        type: array\n        items:\n          type: integer\n          format: int32")
}

func TestOpenAPI_DefaultErrorResponseRefs(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `server S { Foo GET /foo }`)
	assert.Contains(t, out, "default:\n          description: Error\n          content:\n            application/json:\n              schema:\n                $ref: '#/components/schemas/GonzoError'")
	assert.Contains(t, out, "    GonzoError:\n      type: object\n      properties:\n        error:\n          type: string")
}

func TestOpenAPI_EnumValuesQuotedForStrings(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `enum Color string {
  RED = "red"
  BLUE = "blue"
}
server S { Foo GET /foo }`)
	assert.Contains(t, out, "    Color:\n      type: string\n      enum:\n        - blue\n        - red")
}

func TestOpenAPI_EnumIntegerValuesNotQuoted(t *testing.T) {
	out := renderOpenAPIFromSrc(t, `enum Priority int32 {
  LOW = 1
  HIGH = 2
}
server S { Foo GET /foo }`)
	assert.Contains(t, out, "    Priority:\n      type: integer\n      format: int32\n      enum:\n        - 1\n        - 2")
}

func TestOpenAPI_PathsGroupSiblingMethods(t *testing.T) {
	src := `server S {
  GetThing GET /things/{id string}
  DeleteThing DELETE /things/{id string}
}`
	out := renderOpenAPIFromSrc(t, src)
	pathSection := out[strings.Index(out, "  /things/{id}:"):]
	end := strings.Index(pathSection, "\ncomponents:")
	pathSection = pathSection[:end]
	// Both methods under the same path key — neither path appears twice.
	assert.Equal(t, 1, strings.Count(out, "/things/{id}:\n"), "path should be listed once")
	assert.Contains(t, pathSection, "    get:")
	assert.Contains(t, pathSection, "    delete:")
}

func TestOpenAPI_NilAPIReturnsError(t *testing.T) {
	_, err := generator.RenderOpenAPI(nil, "x")
	require.Error(t, err)
}

// TestOpenAPI_GeneratesViaTemplateGenerator verifies the spec is reachable
// via the same NewTemplateGenerator path used by the CLI.
func TestOpenAPI_GeneratesViaTemplateGenerator(t *testing.T) {
	parser := generator.NewParser(`server S { Foo GET /foo }`)
	api, err := parser.Parse()
	require.NoError(t, err)
	g, err := generator.NewTemplateGenerator("languages/openapi/spec/config.yaml")
	require.NoError(t, err)
	results, err := g.Generate(api, "MyAPI")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Contains(t, results, "openapi.yaml")
	assert.Contains(t, results["openapi.yaml"], "title: MyAPI")
	assert.Contains(t, results["openapi.yaml"], "/foo:")
}

// TestOpenAPI_TestServerSnapshot snapshots the full openapi.yaml produced from
// the canonical test_server.api so future template/renderer changes show up
// in diffs.
func TestOpenAPI_TestServerSnapshot(t *testing.T) {
	input, err := fileio.ParseFile("test_data/test_server.api")
	require.NoError(t, err)

	parser := generator.NewParser(string(input))
	api, err := parser.Parse()
	require.NoError(t, err)

	g, err := generator.NewTemplateGenerator("languages/openapi/spec/config.yaml")
	require.NoError(t, err)
	results, err := g.Generate(api, "Gonzo Test API")
	require.NoError(t, err)

	filenames := make([]string, 0, len(results))
	for name := range results {
		filenames = append(filenames, name)
	}
	sort.Strings(filenames)

	for _, name := range filenames {
		t.Run(name, func(t *testing.T) {
			snaps.MatchStandaloneSnapshot(t, results[name])
		})
	}
}
