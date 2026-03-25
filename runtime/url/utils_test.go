package url_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/medubin/gonzo/runtime/url"
	"github.com/stretchr/testify/assert"
)

func TestConvertPathToRegex(t *testing.T) {
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>\w+)/(?P<Test>\w+)/?$`), url.ConvertPathToRegex("/hello/{Message}/{Test}"))
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>[a-z]+)/?$`), url.ConvertPathToRegex("/hello/{Message:[a-z]+}/"))
	assert.Equal(t, regexp.MustCompile(`^/hello/(?P<Message>[a-z]+)/test/(?P<Test>\w+)/?$`), url.ConvertPathToRegex("/hello/{Message:[a-z]+}/test/{Test}/"))
	assert.Equal(t, regexp.MustCompile(`^/hello/test/woot/?$`), url.ConvertPathToRegex("/hello/test/woot"))
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
	ctx := context.WithValue(context.Background(), url.ParamKey{}, params)

	actual := url.GetTypedParamsFromContext[X](ctx)

	assert.Equal(t, params["A"], actual.A)
	assert.Equal(t, params["B"], actual.B)
	assert.Equal(t, params["C"], actual.C)
	assert.Equal(t, "", actual.Empty)
}
