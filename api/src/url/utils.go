package url

import (
	"context"
	"reflect"
	"regexp"
	"strings"
)

type ParamKey = struct{}

func ConvertPathToRegex(path string) *regexp.Regexp {
	path = strings.TrimRight(path, "/")
	// First convert all parameters that are not given an explicit match type. Give it \w+
	unspecifiedConversion := regexp.MustCompile(`{([a-zA-Z0-9]+)}`)
	path = unspecifiedConversion.ReplaceAllString(path, `(?P<$1>\w+)`)

	specifiedConversion := regexp.MustCompile(`{([a-zA-Z0-9]+):(.+)}`)
	path = specifiedConversion.ReplaceAllString(path, `(?P<$1>$2)`)

	// NOTE: ^ means start of string and $ means end. Without these,
	//   we'll still match if the path has content before or after
	//   the expression (/foo/bar/baz would match the "/bar" route).
	return regexp.MustCompile("^" + path + "/?$")
}

func GetTypedParamsFromContext[Params any](ctx context.Context) Params {
	var params Params

	if ctx.Value(ParamKey{}) == nil {
		return params
	}

	for key, value := range ctx.Value(ParamKey{}).(map[string]string) {
		field := reflect.ValueOf(&params).Elem().FieldByName(key)
		if field.IsValid() {
			field.Set(reflect.ValueOf(value))
		}
	}
	return params
}
