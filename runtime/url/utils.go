package url

import (
	"context"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
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

	paramMap := ctx.Value(ParamKey{}).(map[string]string)
	setFieldsFromMap(&params, paramMap, "url")
	return params
}

func GetTypedParamsFromQuery[Params any](query url.Values) Params {
	var params Params

	// Convert url.Values to map[string]string
	paramMap := make(map[string]string)
	for key, values := range query {
		if len(values) > 0 {
			paramMap[key] = values[0]
		}
	}
	
	setFieldsFromMap(&params, paramMap, "json")
	return params
}

// setFieldsFromMap uses reflection to set struct fields based on tag values
func setFieldsFromMap(structPtr interface{}, paramMap map[string]string, tagName string) {
	v := reflect.ValueOf(structPtr).Elem()
	t := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		if !field.CanSet() {
			continue
		}
		
		// Get the tag value (e.g., `url:"role"` or `json:"page"`)
		tag := fieldType.Tag.Get(tagName)
		if tag == "" {
			continue
		}
		
		// Handle json tags that might have ",omitempty"
		if tagName == "json" {
			tag = strings.Split(tag, ",")[0]
		}
		
		// Look for the parameter in the map
		if value, exists := paramMap[tag]; exists {
			setFieldValue(field, value)
		}
	}
}

// setFieldValue sets a struct field value from a string, handling type conversion
func setFieldValue(field reflect.Value, value string) {
	if !field.CanSet() {
		return
	}
	
	fieldType := field.Type()
	
	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		newValue := reflect.New(elemType)
		if setScalarValue(newValue.Elem(), value) {
			field.Set(newValue)
		}
	} else {
		setScalarValue(field, value)
	}
}

// setScalarValue sets a non-pointer field value from a string
func setScalarValue(field reflect.Value, value string) bool {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return true
	case reflect.Int32:
		if i, err := strconv.ParseInt(value, 10, 32); err == nil {
			field.SetInt(i)
			return true
		}
	case reflect.Int64:
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			field.SetInt(i)
			return true
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			field.SetBool(b)
			return true
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(value, 32); err == nil {
			field.SetFloat(f)
			return true
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			field.SetFloat(f)
			return true
		}
	}
	return false
}
