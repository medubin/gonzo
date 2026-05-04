package generator_test

import (
	"strings"
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_FieldDecoratorsParseOntoFields(t *testing.T) {
	api, err := generator.NewParser(`
type R {
  @validation(minLength: 3, maxLength: 32, pattern: "^[a-z]+$")
  required Name string
  @validation(min: 13, max: 120)
  Age int32
}
`).Parse()
	require.NoError(t, err)

	fields := api.Types[0].Fields
	require.Len(t, fields, 2)

	require.Len(t, fields[0].Decorators, 1)
	assert.Equal(t, "validation", fields[0].Decorators[0].Name)
	require.Len(t, fields[0].Decorators[0].Kwargs, 3)

	require.Len(t, fields[1].Decorators, 1)
	require.Len(t, fields[1].Decorators[0].Kwargs, 2)
}

func TestValidation_OpenAPIEmitsConstraints(t *testing.T) {
	api, err := generator.NewParser(`
type R {
  @validation(minLength: 3, maxLength: 32, pattern: "^[a-z]+$", format: "slug")
  required Slug string
  @validation(min: 1, max: 100)
  Count int32
}
server S { Get GET /r returns(R) }
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)

	assert.Contains(t, out, "minLength: 3")
	assert.Contains(t, out, "maxLength: 32")
	assert.Contains(t, out, "pattern: ^[a-z]+$")
	assert.Contains(t, out, "format: slug")
	assert.Contains(t, out, "minimum: 1")
	assert.Contains(t, out, "maximum: 100")
}

func TestValidation_OpenAPISkipsConstraintsOnRefFields(t *testing.T) {
	// `format: email` on a field whose type renders as a $ref must be
	// dropped, since OpenAPI 3.1 forbids most keywords alongside $ref.
	api, err := generator.NewParser(`
type Email string
type R {
  @validation(format: "email")
  required Addr Email
}
server S { Get GET /r returns(R) }
`).Parse()
	require.NoError(t, err)
	// Email is a primitive alias, so it renders as $ref. (renderSchema only
	// treats the literal primitive names as inline.)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)
	// The field should reference Email and NOT carry a sibling format key.
	addrIdx := strings.Index(out, "Addr:")
	require.True(t, addrIdx > 0)
	// Slice from "Addr:" to the next field-or-block to confirm no `format:` line.
	tail := out[addrIdx : addrIdx+200]
	assert.Contains(t, tail, "$ref:")
	assert.NotContains(t, tail, "format: email")
}
