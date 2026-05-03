package url_test

import (
	"context"
	"net/url"
	"regexp"
	"sync"
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

func TestGetTypedParamsFromContext_WrongValueType_ReturnsZero(t *testing.T) {
	type X struct {
		A string `url:"A"`
	}
	ctx := context.WithValue(context.Background(), gonzourl.ParamKey{}, "not-a-map")
	assert.NotPanics(t, func() {
		actual := gonzourl.GetTypedParamsFromContext[X](ctx)
		assert.Equal(t, "", actual.A)
	})
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

// TestGetTypedParamsFromContext_ConcurrentSameType verifies that concurrent
// calls for the same type produce correct results — the key behavioral
// guarantee of the sync.Map cache (no data races, no corruption).
func TestGetTypedParamsFromContext_ConcurrentSameType(t *testing.T) {
	type X struct {
		ID   string `url:"id"`
		Name string `url:"name"`
	}
	params := map[string]string{"id": "42", "name": "alice"}
	ctx := context.WithValue(context.Background(), gonzourl.ParamKey{}, params)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := gonzourl.GetTypedParamsFromContext[X](ctx)
			assert.Equal(t, "42", result.ID)
			assert.Equal(t, "alice", result.Name)
		}()
	}
	wg.Wait()
}
