package url_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/medubin/gonzo/utils/url"
	"github.com/stretchr/testify/assert"
)

func TestConvertPathToRegex(t *testing.T) {
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>\w+)/(?P<Test>\w+)/?$`), url.ConvertPathToRegex("/hello/{Message}/{Test}"))
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>[a-z]+)/?$`), url.ConvertPathToRegex("/hello/{Message:[a-z]+}/"))
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>[a-z]+)/test/(?P<Test>\w+)/?$`), url.ConvertPathToRegex("/hello/{Message:[a-z]+}/test/{Test}/"))
	assert.Equal(t, regexp.MustCompile(`^/hello/test/woot/?$`), url.ConvertPathToRegex("/hello/test/woot"))
}

func TestGetKeys(t *testing.T) {
	assert.Equal(t, []string{"Message", "Test"}, url.GetKeys("/hello/{Message}/{Test}"))
	assert.Equal(t, []string{"Message"}, url.GetKeys("/hello/{Message:[a-z]+}/"))
	assert.Equal(t, []string{"Message", "Test"}, url.GetKeys("/hello/{Message:[a-z]+}/test/{Test}/"))
	assert.Equal(t, []string{}, url.GetKeys("/hello/test/woot"))
}

func TestTestGetKeys(t *testing.T) {
	type X struct {
		A     string
		B     string
		C     string
		Empty string
	}
	params := map[string]string{
		"A": "A1",
		"B": "B2",
		"C": "C3",
	}
	ctx := context.WithValue(context.Background(), url.ParamKey{}, params)

	actual := url.GetTypedParamsFromContext[X](ctx)

	assert.Equal(t, params["A"], actual.A)
	assert.Equal(t, params["B"], actual.B)
	assert.Equal(t, params["C"], actual.C)
	assert.Equal(t, "", actual.Empty)
}
