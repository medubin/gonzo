package url_test

import (
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
