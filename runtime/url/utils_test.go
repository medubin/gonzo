package url_test

import (
	"context"
	"net/url"
	"regexp"
	"testing"

	gonzourl "github.com/medubin/gonzo/runtime/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertPathToRegex(t *testing.T) {
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>\w+)/(?P<Test>\w+)/?$`), gonzourl.ConvertPathToRegex("/hello/{Message}/{Test}"))
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>[a-z]+)/?$`), gonzourl.ConvertPathToRegex("/hello/{Message:[a-z]+}/"))
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>[a-z]+)/test/(?P<Test>\w+)/?$`), gonzourl.ConvertPathToRegex("/hello/{Message:[a-z]+}/test/{Test}/"))
	assert.Equal(t, regexp.MustCompile(`^/hello/test/woot/?$`), gonzourl.ConvertPathToRegex("/hello/test/woot"))
}

func TestTestGetKeys(t *testing.T) {
	type X struct {
		A     string `url:"A"`
		B     string `url:"B"`
		C     string `url:"C"`
		Empty string `url:"Empty"`
	}
	params := map[string]string{
		"A": "A1",
		"B": "B2",
		"C": "C3",
	}
	ctx := context.WithValue(context.Background(), gonzourl.ParamKey{}, params)

	actual := gonzourl.GetTypedParamsFromContext[X](ctx)

	assert.Equal(t, params["A"], actual.A)
	assert.Equal(t, params["B"], actual.B)
	assert.Equal(t, params["C"], actual.C)
	assert.Equal(t, "", actual.Empty)
}

func TestGetTypedParamsFromContext_EmptyContext(t *testing.T) {
	type X struct {
		A string `url:"A"`
	}
	actual := gonzourl.GetTypedParamsFromContext[X](context.Background())
	assert.Equal(t, "", actual.A)
}

func TestGetTypedParamsFromQuery_StringField(t *testing.T) {
	type Q struct {
		Name string `json:"name"`
	}
	q := url.Values{"name": []string{"alice"}}
	result := gonzourl.GetTypedParamsFromQuery[Q](q)
	assert.Equal(t, "alice", result.Name)
}

func TestGetTypedParamsFromQuery_PointerField(t *testing.T) {
	type Q struct {
		Page *string `json:"page,omitempty"`
	}
	q := url.Values{"page": []string{"3"}}
	result := gonzourl.GetTypedParamsFromQuery[Q](q)
	require.NotNil(t, result.Page)
	assert.Equal(t, "3", *result.Page)
}

func TestGetTypedParamsFromQuery_MissingField(t *testing.T) {
	type Q struct {
		Name string `json:"name"`
	}
	q := url.Values{}
	result := gonzourl.GetTypedParamsFromQuery[Q](q)
	assert.Equal(t, "", result.Name)
}

func TestGetTypedParamsFromQuery_MultipleValues_TakesFirst(t *testing.T) {
	type Q struct {
		Tag string `json:"tag"`
	}
	q := url.Values{"tag": []string{"first", "second"}}
	result := gonzourl.GetTypedParamsFromQuery[Q](q)
	assert.Equal(t, "first", result.Tag)
}

func TestGetTypedParamsFromQuery_OmitemptySuffix_Stripped(t *testing.T) {
	type Q struct {
		Limit string `json:"limit,omitempty"`
	}
	q := url.Values{"limit": []string{"10"}}
	result := gonzourl.GetTypedParamsFromQuery[Q](q)
	assert.Equal(t, "10", result.Limit)
}

func TestGetTypedParamsFromContext_Int64Field(t *testing.T) {
	type X struct {
		Count int64 `url:"count"`
	}
	params := map[string]string{"count": "99"}
	ctx := context.WithValue(context.Background(), gonzourl.ParamKey{}, params)
	result := gonzourl.GetTypedParamsFromContext[X](ctx)
	assert.Equal(t, int64(99), result.Count)
}

func TestGetTypedParamsFromContext_BoolField(t *testing.T) {
	type X struct {
		Active bool `url:"active"`
	}
	params := map[string]string{"active": "true"}
	ctx := context.WithValue(context.Background(), gonzourl.ParamKey{}, params)
	result := gonzourl.GetTypedParamsFromContext[X](ctx)
	assert.True(t, result.Active)
}

func TestGetTypedParamsFromContext_InvalidInt_IgnoresField(t *testing.T) {
	type X struct {
		Count int64 `url:"count"`
	}
	params := map[string]string{"count": "not-a-number"}
	ctx := context.WithValue(context.Background(), gonzourl.ParamKey{}, params)
	result := gonzourl.GetTypedParamsFromContext[X](ctx)
	assert.Equal(t, int64(0), result.Count)
}
