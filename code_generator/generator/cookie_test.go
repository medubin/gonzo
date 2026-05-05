package generator_test

import (
	"testing"

	"github.com/medubin/gonzo/code_generator/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCookie_EmittedInOpenAPI(t *testing.T) {
	api, err := generator.NewParser(`
type User { Name string }
server S {
  @cookie("session", required: true, description: "Session token", httpOnly: true, secure: true, sameSite: "Lax")
  @cookie("locale", description: "Locale")
  Login POST /login body(User) returns(User)
}
`).Parse()
	require.NoError(t, err)
	out, err := generator.RenderOpenAPI(api, "T")
	require.NoError(t, err)

	// Required session cookie with description.
	assert.Contains(t, out, "- name: session\n          in: cookie\n          required: true\n          description: Session token\n")
	// Optional locale.
	assert.Contains(t, out, "- name: locale\n          in: cookie\n          description: Locale\n")
	// Write-side attributes (httpOnly etc.) MUST NOT appear in OpenAPI;
	// they're a Go-codegen concern only.
	assert.NotContains(t, out, "httpOnly")
	assert.NotContains(t, out, "sameSite")
}

func TestCookie_TemplateCookieDataPopulated(t *testing.T) {
	// The generator collects cookies into TemplateData.Cookies; this is what
	// drives the cookies.go file. Confirm dedup-by-name works (the same
	// cookie declared on two endpoints folds into one entry).
	api, err := generator.NewParser(`
type X { N string }
server S {
  @cookie("session", httpOnly: true, secure: true)
  Login POST /login body(X) returns(X)

  @cookie("session")
  Logout POST /logout body(X) returns(X)

  @cookie("locale")
  Pref GET /pref returns(X)
}
`).Parse()
	require.NoError(t, err)

	// First declaration (with attributes) should win for write-side metadata.
	cookies := generator.CollectCookies(api)
	require.Len(t, cookies, 2)
	assert.Equal(t, "session", cookies[0].Name)
	assert.Equal(t, "SessionCookieName", cookies[0].ConstName)
	assert.Equal(t, "SetSession", cookies[0].GoSetterName)
	assert.True(t, cookies[0].HttpOnly)
	assert.True(t, cookies[0].Secure)
	assert.True(t, cookies[0].HasWriteAttrs)

	assert.Equal(t, "locale", cookies[1].Name)
	assert.False(t, cookies[1].HasWriteAttrs)
}
